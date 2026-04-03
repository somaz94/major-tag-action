package tagger

import (
	"fmt"
	"os/exec"
	"regexp"
	"strings"
)

// validSHAPattern matches SHA-1 (40 hex) or SHA-256 (64 hex) commit hashes.
var validSHAPattern = regexp.MustCompile(`^[0-9a-f]{40}([0-9a-f]{24})?$`)

// GitRunner defines the interface for executing git commands.
type GitRunner interface {
	Run(args ...string) ([]byte, error)
}

// ExecRunner is the default GitRunner implementation using os/exec.
type ExecRunner struct{}

// Run executes a git command and returns the combined output.
func (r *ExecRunner) Run(args ...string) ([]byte, error) {
	cmd := exec.Command("git", args...)
	return cmd.CombinedOutput()
}

// Git wraps git operations with a pluggable runner.
type Git struct {
	runner GitRunner
}

// NewGit creates a Git instance with the given runner.
func NewGit(runner GitRunner) *Git {
	return &Git{runner: runner}
}

// DefaultGit creates a Git instance using the real exec-based runner.
func DefaultGit() *Git {
	return NewGit(&ExecRunner{})
}

// ConfigureSafeDirectory adds the workspace as a git safe directory.
func (g *Git) ConfigureSafeDirectory(dir string) error {
	_, err := g.runner.Run("config", "--global", "--add", "safe.directory", dir)
	return err
}

// FetchTags fetches all tags from origin.
func (g *Git) FetchTags() error {
	_, err := g.runner.Run("fetch", "--tags", "--force")
	return err
}

// ResolveTagSHA returns the commit SHA for a given tag.
func (g *Git) ResolveTagSHA(tag string) (string, error) {
	out, err := g.runner.Run("rev-list", "-n", "1", tag)
	if err != nil {
		return "", fmt.Errorf("failed to resolve SHA for tag %q: %w", tag, err)
	}
	sha := strings.TrimSpace(string(out))
	if !validSHAPattern.MatchString(sha) {
		return "", fmt.Errorf("%w for tag %q: %q", ErrInvalidSHA, tag, sha)
	}
	return sha, nil
}

// TagExists checks if a tag exists locally.
func (g *Git) TagExists(tag string) bool {
	out, err := g.runner.Run("tag", "-l", tag)
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(out)) == tag
}

// DeleteLocalTag deletes a local tag.
func (g *Git) DeleteLocalTag(tag string) error {
	_, err := g.runner.Run("tag", "-d", tag)
	if err != nil {
		return fmt.Errorf("failed to delete local tag %q: %w", tag, err)
	}
	return nil
}

// DeleteRemoteTag deletes a remote tag.
func (g *Git) DeleteRemoteTag(tag string) error {
	_, err := g.runner.Run("push", "origin", ":refs/tags/"+tag)
	if err != nil {
		return fmt.Errorf("failed to delete remote tag %q: %w", tag, err)
	}
	return nil
}

// CreateTag creates a local tag pointing to a specific commit.
func (g *Git) CreateTag(tag, commitSHA string) error {
	_, err := g.runner.Run("tag", tag, commitSHA)
	if err != nil {
		return fmt.Errorf("failed to create tag %q: %w", tag, err)
	}
	return nil
}

// PushTag pushes a tag to origin.
func (g *Git) PushTag(tag string) error {
	_, err := g.runner.Run("push", "origin", tag)
	if err != nil {
		return fmt.Errorf("failed to push tag %q: %w", tag, err)
	}
	return nil
}

// GetRemoteURL returns the remote origin URL.
func (g *Git) GetRemoteURL() (string, error) {
	out, err := g.runner.Run("remote", "get-url", "origin")
	if err != nil {
		return "", fmt.Errorf("failed to get remote URL: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

// SetRemoteURL updates the remote origin URL.
func (g *Git) SetRemoteURL(url string) error {
	_, err := g.runner.Run("remote", "set-url", "origin", url)
	if err != nil {
		return fmt.Errorf("failed to set remote URL: %w", err)
	}
	return nil
}
