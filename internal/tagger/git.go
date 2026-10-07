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

// credentialInURL matches URL userinfo so a token in https://<token>@github.com/...
// is redacted from git output: a token passed as a plain input is not a
// registered secret, so Actions does not mask it.
var credentialInURL = regexp.MustCompile(`(https?://)[^/@\s]+@`)

// run trims git's output and wraps a failure as "failed to <desc>: <err>[: <output>]"
// (credentials redacted). git's stderr is the only record of why a command failed;
// dropping it is what left the 2026-08-07 `v1` loss undiagnosable.
func (g *Git) run(ctx context.Context, desc string, args ...string) (string, error) {
	out, err := g.runner.Run(ctx, args...)
	if err != nil {
		if detail := strings.TrimSpace(string(out)); detail != "" {
			return "", fmt.Errorf("failed to %s: %w: %s", desc, err, redactCredentials(detail))
		}
		return "", fmt.Errorf("failed to %s: %w", desc, err)
	}
	return strings.TrimSpace(string(out)), nil
}

// redactCredentials replaces any userinfo in a URL with "***".
func redactCredentials(s string) string {
	return credentialInURL.ReplaceAllString(s, "${1}***@")
}

// ConfigureSafeDirectory trusts dir for every child git of this process through
// env config, so no git config file is written.
func (g *Git) ConfigureSafeDirectory(_ context.Context, dir string) error {
	return AddConfigEnv("safe.directory", dir)
}

// FetchTags fetches all tags from origin.
func (g *Git) FetchTags(ctx context.Context) error {
	_, err := g.run(ctx, "fetch tags", "fetch", "--tags", "--force")
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

// CreateTag points a local tag at a specific commit, moving it if it already
// exists. `-f` is what makes this safe to call without deleting first.
func (g *Git) CreateTag(ctx context.Context, tag, commitSHA string) error {
	_, err := g.run(ctx, fmt.Sprintf("create tag %q", tag), "tag", "-f", tag, commitSHA)
	return err
}

// PushTag force-pushes tag to origin as a single remote operation. Delete-then-push
// is two, and a failure between them leaves the tag gone rather than stale, which
// is how `v1` vanished on 2026-08-07. The fully qualified refspec stops the remote
// resolving `v1` to a same-named branch.
func (g *Git) PushTag(ctx context.Context, tag string) error {
	_, err := g.run(ctx, fmt.Sprintf("push tag %q", tag), "push", "--force", "origin",
		fmt.Sprintf("refs/tags/%s:refs/tags/%s", tag, tag))
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
