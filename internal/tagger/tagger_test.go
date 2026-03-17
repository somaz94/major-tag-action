package tagger

import (
	"fmt"
	"os"
	"testing"
)

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

func mockRunner(output []byte, err error) func() {
	original := RunCommand
	RunCommand = func(args ...string) ([]byte, error) {
		return output, err
	}
	return func() { RunCommand = original }
}

func mockRunnerFunc(fn func(args ...string) ([]byte, error)) func() {
	original := RunCommand
	RunCommand = fn
	return func() { RunCommand = original }
}

func TestTagExists(t *testing.T) {
	restore := mockRunner([]byte("v1\n"), nil)
	defer restore()

	if !TagExists("v1") {
		t.Error("expected tag to exist")
	}
}

func TestTagExistsNotFound(t *testing.T) {
	restore := mockRunner([]byte("\n"), nil)
	defer restore()

	if TagExists("v1") {
		t.Error("expected tag not to exist")
	}
}

func TestTagExistsError(t *testing.T) {
	restore := mockRunner(nil, fmt.Errorf("git error"))
	defer restore()

	if TagExists("v1") {
		t.Error("expected false on error")
	}
}

func TestResolveTagSHA(t *testing.T) {
	restore := mockRunner([]byte("abc1234567890\n"), nil)
	defer restore()

	sha, err := ResolveTagSHA("v1.0.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sha != "abc1234567890" {
		t.Errorf("expected abc1234567890, got %s", sha)
	}
}

func TestResolveTagSHAError(t *testing.T) {
	restore := mockRunner(nil, fmt.Errorf("not found"))
	defer restore()

	_, err := ResolveTagSHA("v1.0.0")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestFetchTags(t *testing.T) {
	restore := mockRunner([]byte(""), nil)
	defer restore()

	if err := FetchTags(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestFetchTagsError(t *testing.T) {
	restore := mockRunner(nil, fmt.Errorf("fetch error"))
	defer restore()

	if err := FetchTags(); err == nil {
		t.Fatal("expected error")
	}
}

func TestGetRemoteURL(t *testing.T) {
	restore := mockRunner([]byte("https://github.com/owner/repo.git\n"), nil)
	defer restore()

	url, err := GetRemoteURL()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if url != "https://github.com/owner/repo.git" {
		t.Errorf("expected https://github.com/owner/repo.git, got %s", url)
	}
}

func TestGetRemoteURLError(t *testing.T) {
	restore := mockRunner(nil, fmt.Errorf("no remote"))
	defer restore()

	_, err := GetRemoteURL()
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestUpdateTagNew(t *testing.T) {
	calls := []string{}
	restore := mockRunnerFunc(func(args ...string) ([]byte, error) {
		calls = append(calls, args[0])
		if args[0] == "tag" && len(args) > 1 && args[1] == "-l" {
			return []byte(""), nil // tag doesn't exist
		}
		return []byte(""), nil
	})
	defer restore()

	err := UpdateTag("v1", "abc123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUpdateTagExisting(t *testing.T) {
	restore := mockRunnerFunc(func(args ...string) ([]byte, error) {
		if args[0] == "tag" && len(args) > 1 && args[1] == "-l" {
			return []byte("v1\n"), nil // tag exists
		}
		return []byte(""), nil
	})
	defer restore()

	err := UpdateTag("v1", "abc123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUpdateTagCreateError(t *testing.T) {
	restore := mockRunnerFunc(func(args ...string) ([]byte, error) {
		if args[0] == "tag" && len(args) > 1 && args[1] == "-l" {
			return []byte(""), nil
		}
		if args[0] == "tag" && len(args) > 2 {
			return nil, fmt.Errorf("create error")
		}
		return []byte(""), nil
	})
	defer restore()

	err := UpdateTag("v1", "abc123")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestUpdateTagPushError(t *testing.T) {
	restore := mockRunnerFunc(func(args ...string) ([]byte, error) {
		if args[0] == "tag" && len(args) > 1 && args[1] == "-l" {
			return []byte(""), nil
		}
		if args[0] == "push" {
			return nil, fmt.Errorf("push error")
		}
		return []byte(""), nil
	})
	defer restore()

	err := UpdateTag("v1", "abc123")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestRunSuccess(t *testing.T) {
	restore := mockRunnerFunc(func(args ...string) ([]byte, error) {
		if args[0] == "rev-list" {
			return []byte("abc123def456\n"), nil
		}
		if args[0] == "tag" && len(args) > 1 && args[1] == "-l" {
			return []byte(""), nil
		}
		if args[0] == "remote" {
			return []byte("https://github.com/owner/repo.git\n"), nil
		}
		return []byte(""), nil
	})
	defer restore()

	t.Setenv("GITHUB_WORKSPACE", "/workspace")

	result, err := Run("v1.2.3", true, "token123", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.MajorTag != "v1" {
		t.Errorf("expected v1, got %s", result.MajorTag)
	}
	if result.MinorTag != "" {
		t.Errorf("expected empty minor tag, got %s", result.MinorTag)
	}
	if result.CommitSHA != "abc123def456" {
		t.Errorf("expected abc123def456, got %s", result.CommitSHA)
	}
}

func TestRunWithMinorTag(t *testing.T) {
	restore := mockRunnerFunc(func(args ...string) ([]byte, error) {
		if args[0] == "rev-list" {
			return []byte("abc123\n"), nil
		}
		if args[0] == "tag" && len(args) > 1 && args[1] == "-l" {
			return []byte(""), nil
		}
		if args[0] == "remote" {
			return []byte("https://github.com/owner/repo.git\n"), nil
		}
		return []byte(""), nil
	})
	defer restore()

	t.Setenv("GITHUB_WORKSPACE", "/workspace")

	result, err := Run("v1.2.3", false, "token123", "")
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
	_, err := Run("invalid", true, "", "")
	if err == nil {
		t.Fatal("expected error for invalid tag")
	}
}

func TestRunFetchError(t *testing.T) {
	restore := mockRunnerFunc(func(args ...string) ([]byte, error) {
		if args[0] == "fetch" {
			return nil, fmt.Errorf("fetch error")
		}
		if args[0] == "remote" {
			return []byte("https://github.com/owner/repo.git\n"), nil
		}
		return []byte(""), nil
	})
	defer restore()

	t.Setenv("GITHUB_WORKSPACE", "/workspace")

	_, err := Run("v1.0.0", true, "token", "")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestRunResolveSHAError(t *testing.T) {
	restore := mockRunnerFunc(func(args ...string) ([]byte, error) {
		if args[0] == "rev-list" {
			return nil, fmt.Errorf("not found")
		}
		if args[0] == "remote" {
			return []byte("https://github.com/owner/repo.git\n"), nil
		}
		return []byte(""), nil
	})
	defer restore()

	t.Setenv("GITHUB_WORKSPACE", "/workspace")

	_, err := Run("v1.0.0", true, "token", "")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestConfigureTokenAuthHTTPS(t *testing.T) {
	restore := mockRunnerFunc(func(args ...string) ([]byte, error) {
		if args[0] == "remote" && args[1] == "get-url" {
			return []byte("https://github.com/owner/repo.git\n"), nil
		}
		return []byte(""), nil
	})
	defer restore()

	err := ConfigureAuth("mytoken", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestConfigureTokenAuthSSH(t *testing.T) {
	restore := mockRunnerFunc(func(args ...string) ([]byte, error) {
		if args[0] == "remote" && args[1] == "get-url" {
			return []byte("git@github.com:owner/repo.git\n"), nil
		}
		return []byte(""), nil
	})
	defer restore()

	err := ConfigureAuth("mytoken", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestConfigureAuthNoCredentials(t *testing.T) {
	err := ConfigureAuth("", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestConfigureTokenAuthNonGitHub(t *testing.T) {
	restore := mockRunnerFunc(func(args ...string) ([]byte, error) {
		if args[0] == "remote" && args[1] == "get-url" {
			return []byte("https://gitlab.com/owner/repo.git\n"), nil
		}
		return []byte(""), nil
	})
	defer restore()

	err := ConfigureAuth("mytoken", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestConfigureSSHAuth(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	err := ConfigureAuth("", "fake-ssh-key-content")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestConfigureTokenAuthRemoteError(t *testing.T) {
	restore := mockRunnerFunc(func(args ...string) ([]byte, error) {
		if args[0] == "remote" && args[1] == "get-url" {
			return nil, fmt.Errorf("no remote")
		}
		return []byte(""), nil
	})
	defer restore()

	err := ConfigureAuth("mytoken", "")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestUpdateTagDeleteLocalError(t *testing.T) {
	restore := mockRunnerFunc(func(args ...string) ([]byte, error) {
		if args[0] == "tag" && len(args) > 1 && args[1] == "-l" {
			return []byte("v1\n"), nil
		}
		if args[0] == "tag" && len(args) > 1 && args[1] == "-d" {
			return nil, fmt.Errorf("delete error")
		}
		return []byte(""), nil
	})
	defer restore()

	err := UpdateTag("v1", "abc123")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestRunUpdateMajorTagError(t *testing.T) {
	restore := mockRunnerFunc(func(args ...string) ([]byte, error) {
		if args[0] == "rev-list" {
			return []byte("abc123\n"), nil
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
	defer restore()

	t.Setenv("GITHUB_WORKSPACE", "/workspace")

	_, err := Run("v1.0.0", true, "token", "")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestRunAuthError(t *testing.T) {
	restore := mockRunnerFunc(func(args ...string) ([]byte, error) {
		if args[0] == "remote" && args[1] == "get-url" {
			return nil, fmt.Errorf("no remote")
		}
		return []byte(""), nil
	})
	defer restore()

	t.Setenv("GITHUB_WORKSPACE", "/workspace")

	_, err := Run("v1.0.0", true, "token", "")
	if err == nil {
		t.Fatal("expected error")
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
	restore := mockRunnerFunc(func(args ...string) ([]byte, error) {
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
	defer restore()

	err := UpdateTag("v1", "abc123")
	if err != nil {
		t.Fatalf("expected success even when remote delete fails, got: %v", err)
	}
}

func TestRunMinorTagUpdateError(t *testing.T) {
	callCount := 0
	restore := mockRunnerFunc(func(args ...string) ([]byte, error) {
		if args[0] == "rev-list" {
			return []byte("abc123\n"), nil
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
	defer restore()

	t.Setenv("GITHUB_WORKSPACE", "/workspace")

	_, err := Run("v1.2.3", false, "token", "")
	if err == nil {
		t.Fatal("expected error for minor tag failure")
	}
}

func TestRunDefaultWorkspace(t *testing.T) {
	restore := mockRunnerFunc(func(args ...string) ([]byte, error) {
		if args[0] == "rev-list" {
			return []byte("abc123\n"), nil
		}
		if args[0] == "tag" && len(args) > 1 && args[1] == "-l" {
			return []byte(""), nil
		}
		if args[0] == "remote" {
			return []byte("https://github.com/owner/repo.git\n"), nil
		}
		return []byte(""), nil
	})
	defer restore()

	// No GITHUB_WORKSPACE set - should use default
	t.Setenv("GITHUB_WORKSPACE", "")

	result, err := Run("v1.0.0", true, "token", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.MajorTag != "v1" {
		t.Errorf("expected v1, got %s", result.MajorTag)
	}
}

func TestRunSafeDirectoryError(t *testing.T) {
	restore := mockRunnerFunc(func(args ...string) ([]byte, error) {
		if args[0] == "config" {
			return nil, fmt.Errorf("config error")
		}
		if args[0] == "rev-list" {
			return []byte("abc123\n"), nil
		}
		if args[0] == "tag" && len(args) > 1 && args[1] == "-l" {
			return []byte(""), nil
		}
		if args[0] == "remote" {
			return []byte("https://github.com/owner/repo.git\n"), nil
		}
		return []byte(""), nil
	})
	defer restore()

	t.Setenv("GITHUB_WORKSPACE", "/workspace")

	// Should still succeed - safe directory is just a warning
	result, err := Run("v1.0.0", true, "token", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.MajorTag != "v1" {
		t.Errorf("expected v1, got %s", result.MajorTag)
	}
}

func TestRunWithSSHKey(t *testing.T) {
	restore := mockRunnerFunc(func(args ...string) ([]byte, error) {
		if args[0] == "rev-list" {
			return []byte("abc123\n"), nil
		}
		if args[0] == "tag" && len(args) > 1 && args[1] == "-l" {
			return []byte(""), nil
		}
		return []byte(""), nil
	})
	defer restore()

	t.Setenv("GITHUB_WORKSPACE", "/workspace")
	t.Setenv("HOME", t.TempDir())

	result, err := Run("v2.0.0", true, "", "fake-ssh-key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.MajorTag != "v2" {
		t.Errorf("expected v2, got %s", result.MajorTag)
	}
}
