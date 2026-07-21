package tagger

import (
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
)

// validSHAPattern matches SHA-1 (40 hex) or SHA-256 (64 hex) commit hashes.
var validSHAPattern = regexp.MustCompile(`^[0-9a-f]{40}([0-9a-f]{24})?$`)

// GitRunner defines the interface for executing git commands.
type GitRunner interface {
	Run(ctx context.Context, args ...string) ([]byte, error)
}

// ExecRunner is the default GitRunner implementation using os/exec.
type ExecRunner struct{}

// Run executes a git command and returns the combined output.
func (r *ExecRunner) Run(ctx context.Context, args ...string) ([]byte, error) {
	out, err := exec.CommandContext(ctx, "git", args...).CombinedOutput()
	// Surface the cancellation/timeout cause instead of the opaque
	// "signal: killed" that CombinedOutput reports when ctx fires.
	if err != nil && ctx.Err() != nil {
		return out, ctx.Err()
	}
	return out, err
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

// run executes a git command through the runner, trims the trailing
// whitespace from its output, and wraps any error with a human-readable
// description ("failed to <desc>: <cause>").
func (g *Git) run(ctx context.Context, desc string, args ...string) (string, error) {
	out, err := g.runner.Run(ctx, args...)
	if err != nil {
		return "", fmt.Errorf("failed to %s: %w", desc, err)
	}
	return strings.TrimSpace(string(out)), nil
}

// ConfigureSafeDirectory adds the workspace as a git safe directory.
func (g *Git) ConfigureSafeDirectory(ctx context.Context, dir string) error {
	// Return the raw error unwrapped: the caller logs it as a warning and
	// double-prefixing ("failed to ...: failed to ...") reads poorly there.
	_, err := g.runner.Run(ctx, "config", "--global", "--add", "safe.directory", dir)
	return err
}

// FetchTags fetches all tags from origin.
func (g *Git) FetchTags(ctx context.Context) error {
	// Return the raw error: the caller wraps it with "failed to fetch tags".
	_, err := g.runner.Run(ctx, "fetch", "--tags", "--force")
	return err
}

// ResolveTagSHA returns the commit SHA for a given tag.
func (g *Git) ResolveTagSHA(ctx context.Context, tag string) (string, error) {
	sha, err := g.run(ctx, fmt.Sprintf("resolve SHA for tag %q", tag), "rev-list", "-n", "1", tag)
	if err != nil {
		return "", err
	}
	if !validSHAPattern.MatchString(sha) {
		return "", fmt.Errorf("%w for tag %q: %q", ErrInvalidSHA, tag, sha)
	}
	return sha, nil
}

// TagExists checks if a tag exists locally.
func (g *Git) TagExists(ctx context.Context, tag string) bool {
	out, err := g.run(ctx, fmt.Sprintf("list tag %q", tag), "tag", "-l", tag)
	if err != nil {
		return false
	}
	return out == tag
}

// DeleteLocalTag deletes a local tag.
func (g *Git) DeleteLocalTag(ctx context.Context, tag string) error {
	_, err := g.run(ctx, fmt.Sprintf("delete local tag %q", tag), "tag", "-d", tag)
	return err
}

// DeleteRemoteTag deletes a remote tag.
func (g *Git) DeleteRemoteTag(ctx context.Context, tag string) error {
	_, err := g.run(ctx, fmt.Sprintf("delete remote tag %q", tag), "push", "origin", ":refs/tags/"+tag)
	return err
}

// CreateTag creates a local tag pointing to a specific commit.
func (g *Git) CreateTag(ctx context.Context, tag, commitSHA string) error {
	_, err := g.run(ctx, fmt.Sprintf("create tag %q", tag), "tag", tag, commitSHA)
	return err
}

// PushTag pushes a tag to origin.
func (g *Git) PushTag(ctx context.Context, tag string) error {
	_, err := g.run(ctx, fmt.Sprintf("push tag %q", tag), "push", "origin", tag)
	return err
}

// GetRemoteURL returns the remote origin URL.
func (g *Git) GetRemoteURL(ctx context.Context) (string, error) {
	return g.run(ctx, "get remote URL", "remote", "get-url", "origin")
}

// SetRemoteURL updates the remote origin URL.
func (g *Git) SetRemoteURL(ctx context.Context, url string) error {
	_, err := g.run(ctx, "set remote URL", "remote", "set-url", "origin", url)
	return err
}
