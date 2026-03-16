package tagger

import (
	"fmt"
	"os/exec"
	"strings"
)

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
	return strings.TrimSpace(string(out)), nil
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
	return err
}

// DeleteRemoteTag deletes a remote tag.
func DeleteRemoteTag(tag string) error {
	_, err := RunCommand("push", "origin", ":refs/tags/"+tag)
	return err
}

// CreateTag creates a local tag pointing to a specific commit.
func CreateTag(tag, commitSHA string) error {
	_, err := RunCommand("tag", tag, commitSHA)
	return err
}

// PushTag pushes a tag to origin.
func PushTag(tag string) error {
	_, err := RunCommand("push", "origin", tag)
	return err
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
	return err
}
