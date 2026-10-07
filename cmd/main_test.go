package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/somaz94/major-tag-action/internal/tagger"
)

// mockRunner implements tagger.GitRunner for testing.
type mockRunner struct {
	fn func(args ...string) ([]byte, error)
}

func (m *mockRunner) Run(_ context.Context, args ...string) ([]byte, error) {
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

// captureStdout returns what fn printed to stdout.
func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}
	orig := os.Stdout
	os.Stdout = w
	defer func() { os.Stdout = orig }()

	fn()

	w.Close()
	out, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("failed to read stdout: %v", err)
	}
	return string(out)
}

func TestNotifyShutdownStopIsSilent(t *testing.T) {
	out := captureStdout(t, func() {
		ctx, stop := notifyShutdown()
		stop()
		if ctx.Err() == nil {
			t.Error("expected stop to cancel the context")
		}
	})
	if out != "" {
		t.Errorf("stop logged %q, want nothing", out)
	}
}

func TestNotifyShutdownOnSignal(t *testing.T) {
	out := captureStdout(t, func() {
		ctx, stop := notifyShutdown()
		defer stop()

		self, err := os.FindProcess(os.Getpid())
		if err != nil {
			t.Fatalf("failed to find own process: %v", err)
		}
		if err := self.Signal(syscall.SIGTERM); err != nil {
			t.Fatalf("failed to send SIGTERM: %v", err)
		}

		select {
		case <-ctx.Done():
		case <-time.After(5 * time.Second):
			t.Fatal("context not cancelled after SIGTERM")
		}
	})
	if !strings.Contains(out, "::warning::Received shutdown signal") {
		t.Errorf("expected a shutdown warning, got %q", out)
	}
}
