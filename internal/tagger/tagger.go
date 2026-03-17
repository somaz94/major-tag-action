package tagger

import (
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

// parseVersionParts extracts major and minor version numbers from a semver tag.
func parseVersionParts(tag string) (major, minor string, err error) {
	matches := semverRegex.FindStringSubmatch(tag)
	if matches == nil {
		return "", "", fmt.Errorf("tag %q does not match semver format (expected vX.Y.Z)", tag)
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
func sshDir() string {
	return filepath.Join(os.Getenv("HOME"), ".ssh")
}

// ConfigureAuth sets up git authentication using token or SSH key.
func ConfigureAuth(token, sshKey string) error {
	if sshKey != "" {
		return configureSSHAuth(sshKey)
	}
	if token != "" {
		return configureTokenAuth(token)
	}
	return nil
}

func configureSSHAuth(sshKey string) error {
	sshPath := sshDir()
	if err := os.MkdirAll(sshPath, 0700); err != nil {
		return fmt.Errorf("failed to create .ssh directory: %w", err)
	}

	keyPath := filepath.Join(sshPath, "id_rsa")
	if err := os.WriteFile(keyPath, []byte(sshKey), 0600); err != nil {
		return fmt.Errorf("failed to write SSH key: %w", err)
	}

	knownHostsPath := filepath.Join(sshPath, "known_hosts")
	if err := os.WriteFile(knownHostsPath, []byte(githubSSHRSAKey), 0644); err != nil {
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

func configureTokenAuth(token string) error {
	remoteURL, err := GetRemoteURL()
	if err != nil {
		return err
	}

	if !strings.Contains(remoteURL, "github.com") {
		return nil
	}

	repoPath := extractRepoPath(remoteURL)
	newURL := fmt.Sprintf(tokenAuthURLFormat, token, repoPath)
	return SetRemoteURL(newURL)
}

// UpdateTag deletes the old tag (if exists) and creates a new one pointing to commitSHA.
func UpdateTag(tagName, commitSHA string) error {
	if TagExists(tagName) {
		output.LogInfo(fmt.Sprintf("Deleting existing tag '%s'", tagName))
		if err := DeleteLocalTag(tagName); err != nil {
			return fmt.Errorf("failed to delete local tag %q: %w", tagName, err)
		}
		if err := DeleteRemoteTag(tagName); err != nil {
			output.LogWarning(fmt.Sprintf("Failed to delete remote tag '%s' (may not exist): continuing", tagName))
		}
	}

	output.LogInfo(fmt.Sprintf("Creating tag '%s' pointing to %s", tagName, commitSHA))
	if err := CreateTag(tagName, commitSHA); err != nil {
		return fmt.Errorf("failed to create tag %q: %w", tagName, err)
	}

	if err := PushTag(tagName); err != nil {
		return fmt.Errorf("failed to push tag %q: %w", tagName, err)
	}

	return nil
}

// Run executes the full major tag update workflow.
func Run(tag string, majorOnly bool, token, sshKey string) (*Result, error) {
	majorTag, err := ParseMajorTag(tag)
	if err != nil {
		return nil, err
	}

	output.LogInfo(fmt.Sprintf("Tag: %s", tag))
	output.LogInfo(fmt.Sprintf("Major version tag: %s", majorTag))

	// Configure safe directory
	workspace := os.Getenv("GITHUB_WORKSPACE")
	if workspace == "" {
		workspace = defaultGitHubWorkspace
	}
	if err := ConfigureSafeDirectory(workspace); err != nil {
		output.LogWarning(fmt.Sprintf("Failed to set git safe.directory: %v", err))
	}

	if err := ConfigureAuth(token, sshKey); err != nil {
		return nil, fmt.Errorf("failed to configure authentication: %w", err)
	}

	if err := FetchTags(); err != nil {
		return nil, fmt.Errorf("failed to fetch tags: %w", err)
	}

	commitSHA, err := ResolveTagSHA(tag)
	if err != nil {
		return nil, err
	}
	output.LogInfo(fmt.Sprintf("Commit SHA: %s", commitSHA))

	if err := UpdateTag(majorTag, commitSHA); err != nil {
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
		output.LogInfo(fmt.Sprintf("Minor version tag: %s", minorTag))

		if err := UpdateTag(minorTag, commitSHA); err != nil {
			return nil, err
		}
		result.MinorTag = minorTag
	}

	output.LogInfo(fmt.Sprintf("Successfully updated %s to point to %s (%s)", majorTag, tag, commitSHA))
	return result, nil
}
