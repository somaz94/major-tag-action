package tagger

import (
	"fmt"
	"os/exec"
	"regexp"
	"strings"
)

// validSHAPattern matches SHA-1 (40 hex) or SHA-256 (64 hex) commit hashes.
var validSHAPattern = regexp.MustCompile(`^[0-9a-f]{40}([0-9a-f]{24})?$`)

// RunCommand is a variable to allow mocking in tests.
var RunCommand = func(args ...string) ([]byte, error) {
	cmd := exec.Command("git", args...)
	return cmd.CombinedOutput()
}

// ConfigureSafeDirectory adds the workspace as a git safe directory.
func ConfigureSafeDirectory(dir string) error {
	_, err := RunCommand("config", "--global", "--add", "safe.directory", dir)
	return err
}

// FetchTags fetches all tags from origin.
func FetchTags() error {
	_, err := RunCommand("fetch", "--tags", "--force")
	return err
}

// ResolveTagSHA returns the commit SHA for a given tag.
func ResolveTagSHA(tag string) (string, error) {
	out, err := RunCommand("rev-list", "-n", "1", tag)
	if err != nil {
		return "", fmt.Errorf("failed to resolve SHA for tag %q: %w", tag, err)
	}
	sha := strings.TrimSpace(string(out))
	if !validSHAPattern.MatchString(sha) {
		return "", fmt.Errorf("invalid commit SHA format for tag %q: %q", tag, sha)
	}
	return sha, nil
}

// TagExists checks if a tag exists locally.
func TagExists(tag string) bool {
	out, err := RunCommand("tag", "-l", tag)
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(out)) == tag
}

// DeleteLocalTag deletes a local tag.
func DeleteLocalTag(tag string) error {
	_, err := RunCommand("tag", "-d", tag)
	if err != nil {
		return fmt.Errorf("failed to delete local tag %q: %w", tag, err)
	}
	return nil
}

// DeleteRemoteTag deletes a remote tag.
func DeleteRemoteTag(tag string) error {
	_, err := RunCommand("push", "origin", ":refs/tags/"+tag)
	if err != nil {
		return fmt.Errorf("failed to delete remote tag %q: %w", tag, err)
	}
	return nil
}

// CreateTag creates a local tag pointing to a specific commit.
func CreateTag(tag, commitSHA string) error {
	_, err := RunCommand("tag", tag, commitSHA)
	if err != nil {
		return fmt.Errorf("failed to create tag %q: %w", tag, err)
	}
	return nil
}

// PushTag pushes a tag to origin.
func PushTag(tag string) error {
	_, err := RunCommand("push", "origin", tag)
	if err != nil {
		return fmt.Errorf("failed to push tag %q: %w", tag, err)
	}
	return nil
}


// GetRemoteURL returns the remote origin URL.
func GetRemoteURL() (string, error) {
	out, err := RunCommand("remote", "get-url", "origin")
	if err != nil {
		return "", fmt.Errorf("failed to get remote URL: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

// SetRemoteURL updates the remote origin URL.
func SetRemoteURL(url string) error {
	_, err := RunCommand("remote", "set-url", "origin", url)
	if err != nil {
		return fmt.Errorf("failed to set remote URL: %w", err)
	}
	return nil
}
