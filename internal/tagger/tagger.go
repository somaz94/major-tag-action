package tagger

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/somaz94/major-tag-action/internal/output"
)

var semverRegex = regexp.MustCompile(`^v(\d+)\.(\d+)\.\d+`)

const (
	// githubKnownHosts holds GitHub's published SSH host keys (api.github.com/meta),
	// pinned to the documented fingerprints by TestGitHubKnownHostsFingerprints.
	githubKnownHosts = "github.com ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIOMqqnkVzrm0SdG6UOoqKLsabgH5C9okWi0dh2l9GKJl\n" +
		"github.com ecdsa-sha2-nistp256 AAAAE2VjZHNhLXNoYTItbmlzdHAyNTYAAAAIbmlzdHAyNTYAAABBBEmKSENjQEezOmxkZMy7opKgwFB9nkt5YRrYMjNuG5N87uRgg6CLrbo5wAdT/y6v0mKV0U2w0WZ2YB/++Tpockg=\n" +
		"github.com ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQCj7ndNxQowgcQnjshcLrqPEiiphnt+VTTvDP6mHBL9j1aNUkY4Ue1gvwnGLVlOhGeYrnZaMgRK6+PKCUXaDbC7qtbW8gIkhL7aGCsOr/C56SJMy/BCZfxd1nWzAOxSDPgVsmerOBYfNqltV9/hWCqBywINIR+5dIg6JTJ72pcEpEjcYgXkE2YEFXV1JHnsKgbLWNlhScqb2UmyRkQyytRLtL+38TGxkxCflmO+5Z8CSSNY7GidjMIZ7Q4zMjA2n1nGrlTDkzwDCsw+wqFPGQA179cnfGWOWRVruj16z6XyvxvjJwbz0wQZ75XK5tKSb7FNyeIEs4TT4jk+S4dhPeAUC5y+bDYirYgM4GC7uEnztnZyaVWQ7B381AK4Qdrwt51ZqExKbQpTUNn+EjqoTwvqNj4kqx5QUCI0ThS/YkOxJCXmPUWZbhjpCg56i+2aB6CmK2JGhn57K5mj0MNdBXA4/WnwH6XoPWJzK5Nyu2zB3nAZp+S5hpQs+p1vN1/wsjk=\n"

	// defaultGitHubWorkspace is the default workspace path inside GitHub Actions containers.
	defaultGitHubWorkspace = "/github/workspace"

	// tokenAuthURLFormat is the URL template for token-based authentication.
	tokenAuthURLFormat = "https://x-access-token:%s@github.com/%s.git"

	// sshAuthURLFormat is the URL template for SSH key authentication.
	sshAuthURLFormat = "git@github.com:%s.git"
)

// Result holds the output of the tag update operation.
type Result struct {
	MajorTag  string
	MinorTag  string
	CommitSHA string
}

// Tagger orchestrates the major/minor tag update workflow.
type Tagger struct {
	git *Git
}

// NewTagger creates a Tagger with the given Git instance.
func NewTagger(git *Git) *Tagger {
	return &Tagger{git: git}
}

// DefaultTagger creates a Tagger using the default exec-based git runner.
func DefaultTagger() *Tagger {
	return NewTagger(DefaultGit())
}

// parseVersionParts extracts major and minor version numbers from a semver tag.
func parseVersionParts(tag string) (major, minor string, err error) {
	matches := semverRegex.FindStringSubmatch(tag)
	if matches == nil {
		return "", "", fmt.Errorf("%w: %q", ErrInvalidTag, tag)
	}
	return matches[1], matches[2], nil
}

// ParseMajorTag extracts the major version tag from a semver tag.
// e.g., "v1.2.3" -> "v1"
func ParseMajorTag(tag string) (string, error) {
	major, _, err := parseVersionParts(tag)
	if err != nil {
		return "", err
	}
	return "v" + major, nil
}

// ParseMinorTag extracts the minor version tag from a semver tag.
// e.g., "v1.2.3" -> "v1.2"
func ParseMinorTag(tag string) (string, error) {
	major, minor, err := parseVersionParts(tag)
	if err != nil {
		return "", err
	}
	return "v" + major + "." + minor, nil
}

// sshDir returns the .ssh directory path under HOME.
func sshDir() (string, error) {
	home := os.Getenv("HOME")
	if home == "" {
		return "", fmt.Errorf("HOME environment variable is not set")
	}
	return filepath.Join(home, ".ssh"), nil
}

// ConfigureAuth sets up git authentication using token or SSH key.
func (t *Tagger) ConfigureAuth(ctx context.Context, token, sshKey string) error {
	if sshKey != "" {
		return t.configureSSHAuth(ctx, sshKey)
	}
	if token != "" {
		return t.configureTokenAuth(ctx, token)
	}
	return nil
}

func (t *Tagger) configureSSHAuth(ctx context.Context, sshKey string) error {
	sshPath, err := sshDir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(sshPath, 0700); err != nil {
		return fmt.Errorf("failed to create .ssh directory: %w", err)
	}

	// OpenSSH rejects a key file without a trailing newline ("invalid format"),
	// and a key pasted into a secret often loses it.
	if !strings.HasSuffix(sshKey, "\n") {
		sshKey += "\n"
	}
	keyPath := filepath.Join(sshPath, "id_rsa")
	if err := os.WriteFile(keyPath, []byte(sshKey), 0600); err != nil {
		return fmt.Errorf("failed to write SSH key: %w", err)
	}

	knownHostsPath := filepath.Join(sshPath, "known_hosts")
	if err := os.WriteFile(knownHostsPath, []byte(githubKnownHosts), 0600); err != nil {
		return fmt.Errorf("failed to write known_hosts: %w", err)
	}

	// ssh finds ~/.ssh through the passwd entry, not $HOME: in the container HOME
	// is /github/home but root's home is /root, so both files are named explicitly.
	sshCommand := fmt.Sprintf("ssh -i '%s' -o IdentitiesOnly=yes -o UserKnownHostsFile='%s' -o StrictHostKeyChecking=yes",
		keyPath, knownHostsPath)
	if err := AddConfigEnv("core.sshCommand", sshCommand); err != nil {
		return err
	}

	repoPath, err := t.githubRepoPath(ctx)
	if err != nil {
		return err
	}
	if repoPath == "" {
		return nil
	}
	return t.git.SetRemoteURL(ctx, fmt.Sprintf(sshAuthURLFormat, repoPath))
}

// extractRepoPath extracts the owner/repo path from a GitHub remote URL, in
// https://, ssh:// (with or without a port) or scp-like git@github.com: form.
func extractRepoPath(remoteURL string) string {
	repoPath := strings.TrimSuffix(remoteURL, ".git")

	if u, err := url.Parse(repoPath); err == nil && u.Hostname() == "github.com" && u.Path != "" {
		return strings.TrimPrefix(u.Path, "/")
	}
	if _, rest, ok := strings.Cut(repoPath, "github.com:"); ok {
		return rest
	}

	return repoPath
}

// githubRepoPath returns origin's owner/repo, or "" when origin is not on github.com.
func (t *Tagger) githubRepoPath(ctx context.Context) (string, error) {
	remoteURL, err := t.git.GetRemoteURL(ctx)
	if err != nil {
		return "", err
	}
	if !strings.Contains(remoteURL, "github.com") {
		return "", nil
	}
	return extractRepoPath(remoteURL), nil
}

func (t *Tagger) configureTokenAuth(ctx context.Context, token string) error {
	repoPath, err := t.githubRepoPath(ctx)
	if err != nil {
		return err
	}
	if repoPath == "" {
		return nil
	}
	return t.git.SetRemoteURL(ctx, fmt.Sprintf(tokenAuthURLFormat, token, repoPath))
}

// UpdateTag points tagName at commitSHA, locally and on origin, whether or not
// the tag already exists. It never deletes first (see PushTag): a deleted `v1`
// cannot be restored by this action, whose own release workflow runs it as `@v1`.
func (t *Tagger) UpdateTag(ctx context.Context, tagName, commitSHA string) error {
	output.LogInfo("Pointing tag '" + tagName + "' at " + commitSHA)
	if err := t.git.CreateTag(ctx, tagName, commitSHA); err != nil {
		return err
	}

	return t.git.PushTag(ctx, tagName)
}

// resolveWorkspace returns the configured or default GitHub workspace path.
func resolveWorkspace() string {
	if ws := os.Getenv("GITHUB_WORKSPACE"); ws != "" {
		return ws
	}
	return defaultGitHubWorkspace
}

// Run executes the full major tag update workflow.
func (t *Tagger) Run(ctx context.Context, tag string, majorOnly bool, token, sshKey string) (*Result, error) {
	majorTag, err := ParseMajorTag(tag)
	if err != nil {
		return nil, err
	}

	output.LogInfo("Tag: " + tag)
	output.LogInfo("Major version tag: " + majorTag)

	workspace := resolveWorkspace()
	if err := t.git.ConfigureSafeDirectory(ctx, workspace); err != nil {
		output.LogWarning("Failed to set git safe.directory: " + err.Error())
	}

	if err := t.ConfigureAuth(ctx, token, sshKey); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrAuthFailed, err)
	}

	// Check for cancellation before network operations
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	if err := t.git.FetchTags(ctx); err != nil {
		return nil, err
	}

	commitSHA, err := t.git.ResolveTagSHA(ctx, tag)
	if err != nil {
		return nil, err
	}
	output.LogInfo("Commit SHA: " + commitSHA)

	if err := t.UpdateTag(ctx, majorTag, commitSHA); err != nil {
		return nil, err
	}

	result := &Result{
		MajorTag:  majorTag,
		CommitSHA: commitSHA,
	}

	if !majorOnly {
		minorTag, err := ParseMinorTag(tag)
		if err != nil {
			return nil, err
		}
		output.LogInfo("Minor version tag: " + minorTag)

		if err := t.UpdateTag(ctx, minorTag, commitSHA); err != nil {
			return nil, err
		}
		result.MinorTag = minorTag
	}

	output.LogInfo("Successfully updated " + majorTag + " to point to " + tag + " (" + commitSHA + ")")
	return result, nil
}
