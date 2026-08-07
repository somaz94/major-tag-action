package tagger

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/somaz94/major-tag-action/internal/output"
)

var semverRegex = regexp.MustCompile(`^v(\d+)\.(\d+)\.\d+`)

const (
	// githubSSHRSAKey is the official GitHub SSH RSA host key.
	githubSSHRSAKey = "github.com ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQCj7ndNxQowgcQnjshcLrqPEiiphnt+VTTvDP6mHBL9j1aNUkY4Ue1gvwnGLVlOhGeYrnZaMgRK6+PKCUXaDbC7qtbW8gIkhL7aGCsOr/C56SJMy/BCZfxd1nWzAOxSDPgVsmerOBYfNqltV9/hWCqBywINIR+5dIg6JTJ72pcEpEjcYgXkE2YEFXV1JHnsKgbLWNlhScqb2UmyRkQyytRLtL+38TGxkxCflmO+5Z8CSSNY7GidjMIZ7Q14zNKs=\n"

	// defaultGitHubWorkspace is the default workspace path inside GitHub Actions containers.
	defaultGitHubWorkspace = "/github/workspace"

	// tokenAuthURLFormat is the URL template for token-based authentication.
	tokenAuthURLFormat = "https://x-access-token:%s@github.com/%s.git"
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
		return configureSSHAuth(sshKey)
	}
	if token != "" {
		return t.configureTokenAuth(ctx, token)
	}
	return nil
}

func configureSSHAuth(sshKey string) error {
	sshPath, err := sshDir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(sshPath, 0700); err != nil {
		return fmt.Errorf("failed to create .ssh directory: %w", err)
	}

	keyPath := filepath.Join(sshPath, "id_rsa")
	if err := os.WriteFile(keyPath, []byte(sshKey), 0600); err != nil {
		return fmt.Errorf("failed to write SSH key: %w", err)
	}

	knownHostsPath := filepath.Join(sshPath, "known_hosts")
	if err := os.WriteFile(knownHostsPath, []byte(githubSSHRSAKey), 0600); err != nil {
		return fmt.Errorf("failed to write known_hosts: %w", err)
	}

	return nil
}

// extractRepoPath extracts the owner/repo path from a GitHub remote URL.
func extractRepoPath(remoteURL string) string {
	repoPath := strings.TrimSuffix(remoteURL, ".git")

	if strings.HasPrefix(repoPath, "https://") {
		parts := strings.SplitN(repoPath, "github.com/", 2)
		if len(parts) == 2 {
			return parts[1]
		}
	} else if strings.Contains(repoPath, "github.com:") {
		parts := strings.SplitN(repoPath, "github.com:", 2)
		if len(parts) == 2 {
			return parts[1]
		}
	}

	return repoPath
}

func (t *Tagger) configureTokenAuth(ctx context.Context, token string) error {
	remoteURL, err := t.git.GetRemoteURL(ctx)
	if err != nil {
		return err
	}

	if !strings.Contains(remoteURL, "github.com") {
		return nil
	}

	repoPath := extractRepoPath(remoteURL)
	newURL := fmt.Sprintf(tokenAuthURLFormat, token, repoPath)
	return t.git.SetRemoteURL(ctx, newURL)
}

// UpdateTag points tagName at commitSHA, locally and on origin, whether or not
// the tag already exists.
//
// It does NOT delete the tag first. The remote side is a single force push, so
// a failure leaves the tag pointing where it did before instead of leaving it
// deleted — the state that took this action's own release pipeline down, since
// the workflow that recreates `v1` is itself referenced as `@v1`.
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

	// Configure safe directory
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
		return nil, fmt.Errorf("failed to fetch tags: %w", err)
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
