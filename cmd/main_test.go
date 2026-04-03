package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/somaz94/major-tag-action/internal/tagger"
)

// mockRunner implements tagger.GitRunner for testing.
type mockRunner struct {
	fn func(args ...string) ([]byte, error)
}

func (m *mockRunner) Run(args ...string) ([]byte, error) {
	return m.fn(args...)
}

func newTestTagger(fn func(args ...string) ([]byte, error)) *tagger.Tagger {
	return tagger.NewTagger(tagger.NewGit(&mockRunner{fn: fn}))
}

func TestRunEmptyTag(t *testing.T) {
	t.Setenv("INPUT_TAG", "")

	ctx := context.Background()
	tgr := tagger.DefaultTagger()
	err := run(ctx, tgr)
	if err == nil {
		t.Fatal("expected error for empty tag")
	}
}

func TestRunSuccess(t *testing.T) {
	tgr := newTestTagger(func(args ...string) ([]byte, error) {
		if args[0] == "rev-list" {
			return []byte("abc123def456abc123def456abc123def456abc1\n"), nil
		}
		if args[0] == "tag" && len(args) > 1 && args[1] == "-l" {
			return []byte(""), nil
		}
		if args[0] == "remote" {
			return []byte("https://github.com/owner/repo.git\n"), nil
		}
		return []byte(""), nil
	})

	tmpDir := t.TempDir()
	t.Setenv("INPUT_TAG", "v1.2.3")
	t.Setenv("INPUT_MAJOR_ONLY", "true")
	t.Setenv("INPUT_GITHUB_TOKEN", "token")
	t.Setenv("INPUT_SSH_KEY", "")
	t.Setenv("GITHUB_WORKSPACE", tmpDir)

	ctx := context.Background()
	err := run(ctx, tgr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunFailure(t *testing.T) {
	tgr := newTestTagger(func(args ...string) ([]byte, error) {
		if args[0] == "remote" && args[1] == "get-url" {
			return nil, fmt.Errorf("no remote")
		}
		return []byte(""), nil
	})

	t.Setenv("INPUT_TAG", "v1.0.0")
	t.Setenv("INPUT_MAJOR_ONLY", "true")
	t.Setenv("INPUT_GITHUB_TOKEN", "token")
	t.Setenv("INPUT_SSH_KEY", "")
	t.Setenv("GITHUB_WORKSPACE", t.TempDir())

	ctx := context.Background()
	err := run(ctx, tgr)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestRunWithGitHubOutput(t *testing.T) {
	tgr := newTestTagger(func(args ...string) ([]byte, error) {
		if args[0] == "rev-list" {
			return []byte("abc123def456abc123def456abc123def456abc1\n"), nil
		}
		if args[0] == "tag" && len(args) > 1 && args[1] == "-l" {
			return []byte(""), nil
		}
		if args[0] == "remote" {
			return []byte("https://github.com/owner/repo.git\n"), nil
		}
		return []byte(""), nil
	})

	tmpDir := t.TempDir()
	outputFile := filepath.Join(tmpDir, "github_output")
	os.WriteFile(outputFile, []byte{}, 0644)

	t.Setenv("INPUT_TAG", "v1.2.3")
	t.Setenv("INPUT_MAJOR_ONLY", "true")
	t.Setenv("INPUT_GITHUB_TOKEN", "token")
	t.Setenv("INPUT_SSH_KEY", "")
	t.Setenv("GITHUB_WORKSPACE", tmpDir)
	t.Setenv("GITHUB_OUTPUT", outputFile)

	ctx := context.Background()
	err := run(ctx, tgr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, _ := os.ReadFile(outputFile)
	if len(data) == 0 {
		t.Error("expected GITHUB_OUTPUT to have content")
	}
}

func TestRunCancelled(t *testing.T) {
	t.Setenv("INPUT_TAG", "v1.0.0")

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	tgr := tagger.DefaultTagger()
	err := run(ctx, tgr)
	if err == nil {
		t.Fatal("expected error for cancelled context")
	}
	if err.Error() != "cancelled" {
		t.Errorf("expected 'cancelled' error, got %q", err.Error())
	}
}
