package tagger

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"
)

// MockRunner implements GitRunner for testing.
type MockRunner struct {
	Fn func(args ...string) ([]byte, error)
}

func (m *MockRunner) Run(args ...string) ([]byte, error) {
	return m.Fn(args...)
}

func newMockGit(fn func(args ...string) ([]byte, error)) *Git {
	return NewGit(&MockRunner{Fn: fn})
}

func newMockTagger(fn func(args ...string) ([]byte, error)) *Tagger {
	return NewTagger(newMockGit(fn))
}

func staticMockGit(output []byte, err error) *Git {
	return newMockGit(func(args ...string) ([]byte, error) {
		return output, err
	})
}

func TestParseMajorTag(t *testing.T) {
	tests := []struct {
		input    string
		expected string
		wantErr  bool
	}{
		{"v1.2.3", "v1", false},
		{"v0.1.0", "v0", false},
		{"v12.34.56", "v12", false},
		{"v1.2.3-rc1", "v1", false},
		{"invalid", "", true},
		{"1.2.3", "", true},
		{"v1", "", true},
		{"v1.2", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result, err := ParseMajorTag(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error for %q, got %q", tt.input, result)
				}
				if !errors.Is(err, ErrInvalidTag) {
					t.Errorf("expected ErrInvalidTag, got: %v", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error for %q: %v", tt.input, err)
			}
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestParseMinorTag(t *testing.T) {
	tests := []struct {
		input    string
		expected string
		wantErr  bool
	}{
		{"v1.2.3", "v1.2", false},
		{"v0.1.0", "v0.1", false},
		{"v12.34.56", "v12.34", false},
		{"invalid", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result, err := ParseMinorTag(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error for %q, got %q", tt.input, result)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error for %q: %v", tt.input, err)
			}
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestTagExists(t *testing.T) {
	git := staticMockGit([]byte("v1\n"), nil)

	if !git.TagExists("v1") {
		t.Error("expected tag to exist")
	}
}

func TestTagExistsNotFound(t *testing.T) {
	git := staticMockGit([]byte("\n"), nil)

	if git.TagExists("v1") {
		t.Error("expected tag not to exist")
	}
}

func TestTagExistsError(t *testing.T) {
	git := staticMockGit(nil, fmt.Errorf("git error"))

	if git.TagExists("v1") {
		t.Error("expected false on error")
	}
}

func TestResolveTagSHA(t *testing.T) {
	validSHA := "abc1234567890abc1234567890abc1234567890a"
	git := staticMockGit([]byte(validSHA+"\n"), nil)

	sha, err := git.ResolveTagSHA("v1.0.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sha != validSHA {
		t.Errorf("expected %s, got %s", validSHA, sha)
	}
}

func TestResolveTagSHAInvalidFormat(t *testing.T) {
	git := staticMockGit([]byte("not-a-valid-sha\n"), nil)

	_, err := git.ResolveTagSHA("v1.0.0")
	if err == nil {
		t.Fatal("expected error for invalid SHA format")
	}
	if !errors.Is(err, ErrInvalidSHA) {
		t.Errorf("expected ErrInvalidSHA, got: %v", err)
	}
}

func TestResolveTagSHAError(t *testing.T) {
	git := staticMockGit(nil, fmt.Errorf("not found"))

	_, err := git.ResolveTagSHA("v1.0.0")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestFetchTags(t *testing.T) {
	git := staticMockGit([]byte(""), nil)

	if err := git.FetchTags(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestFetchTagsError(t *testing.T) {
	git := staticMockGit(nil, fmt.Errorf("fetch error"))

	if err := git.FetchTags(); err == nil {
		t.Fatal("expected error")
	}
}

func TestGetRemoteURL(t *testing.T) {
	git := staticMockGit([]byte("https://github.com/owner/repo.git\n"), nil)

	url, err := git.GetRemoteURL()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if url != "https://github.com/owner/repo.git" {
		t.Errorf("expected https://github.com/owner/repo.git, got %s", url)
	}
}

func TestGetRemoteURLError(t *testing.T) {
	git := staticMockGit(nil, fmt.Errorf("no remote"))

	_, err := git.GetRemoteURL()
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestUpdateTagNew(t *testing.T) {
	tgr := newMockTagger(func(args ...string) ([]byte, error) {
		if args[0] == "tag" && len(args) > 1 && args[1] == "-l" {
			return []byte(""), nil // tag doesn't exist
		}
		return []byte(""), nil
	})

	err := tgr.UpdateTag("v1", "abc123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUpdateTagExisting(t *testing.T) {
	tgr := newMockTagger(func(args ...string) ([]byte, error) {
		if args[0] == "tag" && len(args) > 1 && args[1] == "-l" {
			return []byte("v1\n"), nil // tag exists
		}
		return []byte(""), nil
	})

	err := tgr.UpdateTag("v1", "abc123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUpdateTagCreateError(t *testing.T) {
	tgr := newMockTagger(func(args ...string) ([]byte, error) {
		if args[0] == "tag" && len(args) > 1 && args[1] == "-l" {
			return []byte(""), nil
		}
		if args[0] == "tag" && len(args) > 2 {
			return nil, fmt.Errorf("create error")
		}
		return []byte(""), nil
	})

	err := tgr.UpdateTag("v1", "abc123")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestUpdateTagPushError(t *testing.T) {
	tgr := newMockTagger(func(args ...string) ([]byte, error) {
		if args[0] == "tag" && len(args) > 1 && args[1] == "-l" {
			return []byte(""), nil
		}
		if args[0] == "push" {
			return nil, fmt.Errorf("push error")
		}
		return []byte(""), nil
	})

	err := tgr.UpdateTag("v1", "abc123")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestRunSuccess(t *testing.T) {
	tgr := newMockTagger(func(args ...string) ([]byte, error) {
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

	t.Setenv("GITHUB_WORKSPACE", "/workspace")

	result, err := tgr.Run(context.Background(), "v1.2.3", true, "token123", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.MajorTag != "v1" {
		t.Errorf("expected v1, got %s", result.MajorTag)
	}
	if result.MinorTag != "" {
		t.Errorf("expected empty minor tag, got %s", result.MinorTag)
	}
	if result.CommitSHA != "abc123def456abc123def456abc123def456abc1" {
		t.Errorf("expected abc123def456abc123def456abc123def456abc1, got %s", result.CommitSHA)
	}
}

func TestRunWithMinorTag(t *testing.T) {
	tgr := newMockTagger(func(args ...string) ([]byte, error) {
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

	t.Setenv("GITHUB_WORKSPACE", "/workspace")

	result, err := tgr.Run(context.Background(), "v1.2.3", false, "token123", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.MajorTag != "v1" {
		t.Errorf("expected v1, got %s", result.MajorTag)
	}
	if result.MinorTag != "v1.2" {
		t.Errorf("expected v1.2, got %s", result.MinorTag)
	}
}

func TestRunInvalidTag(t *testing.T) {
	tgr := newMockTagger(func(args ...string) ([]byte, error) {
		return []byte(""), nil
	})

	_, err := tgr.Run(context.Background(), "invalid", true, "", "")
	if err == nil {
		t.Fatal("expected error for invalid tag")
	}
	if !errors.Is(err, ErrInvalidTag) {
		t.Errorf("expected ErrInvalidTag, got: %v", err)
	}
}

func TestRunFetchError(t *testing.T) {
	tgr := newMockTagger(func(args ...string) ([]byte, error) {
		if args[0] == "fetch" {
			return nil, fmt.Errorf("fetch error")
		}
		if args[0] == "remote" {
			return []byte("https://github.com/owner/repo.git\n"), nil
		}
		return []byte(""), nil
	})

	t.Setenv("GITHUB_WORKSPACE", "/workspace")

	_, err := tgr.Run(context.Background(), "v1.0.0", true, "token", "")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestRunResolveSHAError(t *testing.T) {
	tgr := newMockTagger(func(args ...string) ([]byte, error) {
		if args[0] == "rev-list" {
			return nil, fmt.Errorf("not found")
		}
		if args[0] == "remote" {
			return []byte("https://github.com/owner/repo.git\n"), nil
		}
		return []byte(""), nil
	})

	t.Setenv("GITHUB_WORKSPACE", "/workspace")

	_, err := tgr.Run(context.Background(), "v1.0.0", true, "token", "")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestConfigureTokenAuthHTTPS(t *testing.T) {
	tgr := newMockTagger(func(args ...string) ([]byte, error) {
		if args[0] == "remote" && args[1] == "get-url" {
			return []byte("https://github.com/owner/repo.git\n"), nil
		}
		return []byte(""), nil
	})

	err := tgr.ConfigureAuth("mytoken", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestConfigureTokenAuthSSH(t *testing.T) {
	tgr := newMockTagger(func(args ...string) ([]byte, error) {
		if args[0] == "remote" && args[1] == "get-url" {
			return []byte("git@github.com:owner/repo.git\n"), nil
		}
		return []byte(""), nil
	})

	err := tgr.ConfigureAuth("mytoken", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestConfigureAuthNoCredentials(t *testing.T) {
	tgr := newMockTagger(func(args ...string) ([]byte, error) {
		return []byte(""), nil
	})

	err := tgr.ConfigureAuth("", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestConfigureTokenAuthNonGitHub(t *testing.T) {
	tgr := newMockTagger(func(args ...string) ([]byte, error) {
		if args[0] == "remote" && args[1] == "get-url" {
			return []byte("https://gitlab.com/owner/repo.git\n"), nil
		}
		return []byte(""), nil
	})

	err := tgr.ConfigureAuth("mytoken", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestConfigureSSHAuth(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	tgr := newMockTagger(func(args ...string) ([]byte, error) {
		return []byte(""), nil
	})

	err := tgr.ConfigureAuth("", "fake-ssh-key-content")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestConfigureTokenAuthRemoteError(t *testing.T) {
	tgr := newMockTagger(func(args ...string) ([]byte, error) {
		if args[0] == "remote" && args[1] == "get-url" {
			return nil, fmt.Errorf("no remote")
		}
		return []byte(""), nil
	})

	err := tgr.ConfigureAuth("mytoken", "")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestUpdateTagDeleteLocalError(t *testing.T) {
	tgr := newMockTagger(func(args ...string) ([]byte, error) {
		if args[0] == "tag" && len(args) > 1 && args[1] == "-l" {
			return []byte("v1\n"), nil
		}
		if args[0] == "tag" && len(args) > 1 && args[1] == "-d" {
			return nil, fmt.Errorf("delete error")
		}
		return []byte(""), nil
	})

	err := tgr.UpdateTag("v1", "abc123")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestRunUpdateMajorTagError(t *testing.T) {
	tgr := newMockTagger(func(args ...string) ([]byte, error) {
		if args[0] == "rev-list" {
			return []byte("abc123def456abc123def456abc123def456abc1\n"), nil
		}
		if args[0] == "tag" && len(args) > 1 && args[1] == "-l" {
			return []byte(""), nil
		}
		if args[0] == "tag" && len(args) > 2 {
			return nil, fmt.Errorf("tag create error")
		}
		if args[0] == "remote" {
			return []byte("https://github.com/owner/repo.git\n"), nil
		}
		return []byte(""), nil
	})

	t.Setenv("GITHUB_WORKSPACE", "/workspace")

	_, err := tgr.Run(context.Background(), "v1.0.0", true, "token", "")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestRunAuthError(t *testing.T) {
	tgr := newMockTagger(func(args ...string) ([]byte, error) {
		if args[0] == "remote" && args[1] == "get-url" {
			return nil, fmt.Errorf("no remote")
		}
		return []byte(""), nil
	})

	t.Setenv("GITHUB_WORKSPACE", "/workspace")

	_, err := tgr.Run(context.Background(), "v1.0.0", true, "token", "")
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, ErrAuthFailed) {
		t.Errorf("expected ErrAuthFailed, got: %v", err)
	}
}

func TestConfigureSSHAuthBadHome(t *testing.T) {
	// HOME pointing to a read-only file to trigger MkdirAll error
	tmpDir := t.TempDir()
	badPath := tmpDir + "/blocked"
	os.WriteFile(badPath, []byte("x"), 0444) // create file, not dir
	t.Setenv("HOME", badPath)

	err := configureSSHAuth("fake-key")
	if err == nil {
		t.Fatal("expected error for bad HOME path")
	}
}

func TestConfigureSSHAuthWriteKeyError(t *testing.T) {
	// .ssh dir exists but key path is a directory → WriteFile fails
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	os.MkdirAll(tmpDir+"/.ssh/id_rsa", 0700) // create dir where file should be

	err := configureSSHAuth("fake-key")
	if err == nil {
		t.Fatal("expected error for write key failure")
	}
}

func TestConfigureSSHAuthWriteKnownHostsError(t *testing.T) {
	// key write succeeds, but known_hosts path is a directory → WriteFile fails
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	os.MkdirAll(tmpDir+"/.ssh/known_hosts", 0700) // create dir where file should be

	err := configureSSHAuth("fake-key")
	if err == nil {
		t.Fatal("expected error for write known_hosts failure")
	}
}

func TestExtractRepoPathHTTPS(t *testing.T) {
	result := extractRepoPath("https://github.com/owner/repo.git")
	if result != "owner/repo" {
		t.Errorf("expected owner/repo, got %s", result)
	}
}

func TestExtractRepoPathSSH(t *testing.T) {
	result := extractRepoPath("git@github.com:owner/repo.git")
	if result != "owner/repo" {
		t.Errorf("expected owner/repo, got %s", result)
	}
}

func TestExtractRepoPathPlain(t *testing.T) {
	// URL without github.com prefix patterns → returns as-is minus .git
	result := extractRepoPath("https://gitlab.com/owner/repo.git")
	if result != "https://gitlab.com/owner/repo" {
		t.Errorf("unexpected result: %s", result)
	}
}

func TestExtractRepoPathNoSplit(t *testing.T) {
	// https URL with github.com but no slash after it
	result := extractRepoPath("https://github.com")
	if result != "https://github.com" {
		t.Errorf("unexpected result: %s", result)
	}
}

func TestUpdateTagDeleteRemoteErrorContinues(t *testing.T) {
	// When tag exists, delete local succeeds, delete remote fails (should continue)
	tgr := newMockTagger(func(args ...string) ([]byte, error) {
		if args[0] == "tag" && len(args) > 1 && args[1] == "-l" {
			return []byte("v1\n"), nil
		}
		if args[0] == "tag" && len(args) > 1 && args[1] == "-d" {
			return []byte(""), nil
		}
		if args[0] == "push" && len(args) > 1 && args[1] == "origin" && args[2] == ":refs/tags/v1" {
			return nil, fmt.Errorf("remote delete failed")
		}
		return []byte(""), nil
	})

	err := tgr.UpdateTag("v1", "abc123")
	if err != nil {
		t.Fatalf("expected success even when remote delete fails, got: %v", err)
	}
}

func TestRunMinorTagUpdateError(t *testing.T) {
	callCount := 0
	tgr := newMockTagger(func(args ...string) ([]byte, error) {
		if args[0] == "rev-list" {
			return []byte("abc123def456abc123def456abc123def456abc1\n"), nil
		}
		if args[0] == "tag" && len(args) > 1 && args[1] == "-l" {
			return []byte(""), nil
		}
		if args[0] == "tag" && len(args) > 2 {
			callCount++
			if callCount > 1 {
				// Fail on second tag create (minor tag)
				return nil, fmt.Errorf("minor tag create error")
			}
			return []byte(""), nil
		}
		if args[0] == "remote" {
			return []byte("https://github.com/owner/repo.git\n"), nil
		}
		return []byte(""), nil
	})

	t.Setenv("GITHUB_WORKSPACE", "/workspace")

	_, err := tgr.Run(context.Background(), "v1.2.3", false, "token", "")
	if err == nil {
		t.Fatal("expected error for minor tag failure")
	}
}

func TestRunDefaultWorkspace(t *testing.T) {
	tgr := newMockTagger(func(args ...string) ([]byte, error) {
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

	// No GITHUB_WORKSPACE set - should use default
	t.Setenv("GITHUB_WORKSPACE", "")

	result, err := tgr.Run(context.Background(), "v1.0.0", true, "token", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.MajorTag != "v1" {
		t.Errorf("expected v1, got %s", result.MajorTag)
	}
}

func TestRunSafeDirectoryError(t *testing.T) {
	tgr := newMockTagger(func(args ...string) ([]byte, error) {
		if args[0] == "config" {
			return nil, fmt.Errorf("config error")
		}
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

	t.Setenv("GITHUB_WORKSPACE", "/workspace")

	// Should still succeed - safe directory is just a warning
	result, err := tgr.Run(context.Background(), "v1.0.0", true, "token", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.MajorTag != "v1" {
		t.Errorf("expected v1, got %s", result.MajorTag)
	}
}

func TestRunWithSSHKey(t *testing.T) {
	tgr := newMockTagger(func(args ...string) ([]byte, error) {
		if args[0] == "rev-list" {
			return []byte("abc123def456abc123def456abc123def456abc1\n"), nil
		}
		if args[0] == "tag" && len(args) > 1 && args[1] == "-l" {
			return []byte(""), nil
		}
		return []byte(""), nil
	})

	t.Setenv("GITHUB_WORKSPACE", "/workspace")
	t.Setenv("HOME", t.TempDir())

	result, err := tgr.Run(context.Background(), "v2.0.0", true, "", "fake-ssh-key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.MajorTag != "v2" {
		t.Errorf("expected v2, got %s", result.MajorTag)
	}
}

func TestConfigureSSHAuthEmptyHome(t *testing.T) {
	t.Setenv("HOME", "")

	err := configureSSHAuth("fake-key")
	if err == nil {
		t.Fatal("expected error for empty HOME")
	}
	if !strings.Contains(err.Error(), "HOME environment variable") {
		t.Errorf("expected HOME error, got: %v", err)
	}
}

func TestValidSHAPattern(t *testing.T) {
	tests := []struct {
		input string
		valid bool
	}{
		{"abc123def456abc123def456abc123def456abc1", true},                         // 40 hex (SHA-1)
		{"abc123def456abc123def456abc123def456abc1aabbccdd00112233aabbccdd", true}, // 64 hex (SHA-256)
		{"not-a-sha", false},
		{"ABC123DEF456ABC123DEF456ABC123DEF456ABC1", false}, // uppercase
		{"abc123", false}, // too short
		{"", false},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if validSHAPattern.MatchString(tt.input) != tt.valid {
				t.Errorf("validSHAPattern(%q) = %v, want %v", tt.input, !tt.valid, tt.valid)
			}
		})
	}
}

func TestRunContextCancelled(t *testing.T) {
	tgr := newMockTagger(func(args ...string) ([]byte, error) {
		if args[0] == "remote" {
			return []byte("https://github.com/owner/repo.git\n"), nil
		}
		return []byte(""), nil
	})

	t.Setenv("GITHUB_WORKSPACE", "/workspace")

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := tgr.Run(ctx, "v1.0.0", true, "token", "")
	if err == nil {
		t.Fatal("expected error for cancelled context")
	}
}

func TestNewGitAndDefaultGit(t *testing.T) {
	// Test NewGit with custom runner
	mock := &MockRunner{Fn: func(args ...string) ([]byte, error) {
		return []byte("ok"), nil
	}}
	git := NewGit(mock)
	if git == nil {
		t.Fatal("expected non-nil Git")
	}

	// Test DefaultGit creates ExecRunner
	defaultGit := DefaultGit()
	if defaultGit == nil {
		t.Fatal("expected non-nil default Git")
	}
}

func TestDefaultTagger(t *testing.T) {
	tgr := DefaultTagger()
	if tgr == nil {
		t.Fatal("expected non-nil default Tagger")
	}
}

func TestDeleteLocalTag(t *testing.T) {
	git := staticMockGit([]byte(""), nil)
	if err := git.DeleteLocalTag("v1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDeleteLocalTagError(t *testing.T) {
	git := staticMockGit(nil, fmt.Errorf("delete error"))
	err := git.DeleteLocalTag("v1")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestDeleteRemoteTag(t *testing.T) {
	git := staticMockGit([]byte(""), nil)
	if err := git.DeleteRemoteTag("v1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDeleteRemoteTagError(t *testing.T) {
	git := staticMockGit(nil, fmt.Errorf("push error"))
	err := git.DeleteRemoteTag("v1")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestCreateTag(t *testing.T) {
	git := staticMockGit([]byte(""), nil)
	if err := git.CreateTag("v1", "abc123"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCreateTagError(t *testing.T) {
	git := staticMockGit(nil, fmt.Errorf("tag error"))
	err := git.CreateTag("v1", "abc123")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestPushTag(t *testing.T) {
	git := staticMockGit([]byte(""), nil)
	if err := git.PushTag("v1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestPushTagError(t *testing.T) {
	git := staticMockGit(nil, fmt.Errorf("push error"))
	err := git.PushTag("v1")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestSetRemoteURL(t *testing.T) {
	git := staticMockGit([]byte(""), nil)
	if err := git.SetRemoteURL("https://example.com"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSetRemoteURLError(t *testing.T) {
	git := staticMockGit(nil, fmt.Errorf("set-url error"))
	err := git.SetRemoteURL("https://example.com")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestConfigureSafeDirectory(t *testing.T) {
	git := staticMockGit([]byte(""), nil)
	if err := git.ConfigureSafeDirectory("/workspace"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestConfigureSafeDirectoryError(t *testing.T) {
	git := staticMockGit(nil, fmt.Errorf("config error"))
	err := git.ConfigureSafeDirectory("/workspace")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestResolveWorkspace(t *testing.T) {
	t.Setenv("GITHUB_WORKSPACE", "/custom/workspace")
	if ws := resolveWorkspace(); ws != "/custom/workspace" {
		t.Errorf("expected /custom/workspace, got %s", ws)
	}

	t.Setenv("GITHUB_WORKSPACE", "")
	if ws := resolveWorkspace(); ws != defaultGitHubWorkspace {
		t.Errorf("expected %s, got %s", defaultGitHubWorkspace, ws)
	}
}
