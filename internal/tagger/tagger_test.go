package tagger

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"slices"
	"strings"
	"testing"
)

// MockRunner implements GitRunner for testing.
type MockRunner struct {
	Fn func(args ...string) ([]byte, error)
}

func (m *MockRunner) Run(_ context.Context, args ...string) ([]byte, error) {
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

func TestResolveTagSHA(t *testing.T) {
	validSHA := "abc1234567890abc1234567890abc1234567890a"
	git := staticMockGit([]byte(validSHA+"\n"), nil)

	sha, err := git.ResolveTagSHA(context.Background(), "v1.0.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sha != validSHA {
		t.Errorf("expected %s, got %s", validSHA, sha)
	}
}

func TestResolveTagSHAInvalidFormat(t *testing.T) {
	git := staticMockGit([]byte("not-a-valid-sha\n"), nil)

	_, err := git.ResolveTagSHA(context.Background(), "v1.0.0")
	if err == nil {
		t.Fatal("expected error for invalid SHA format")
	}
	if !errors.Is(err, ErrInvalidSHA) {
		t.Errorf("expected ErrInvalidSHA, got: %v", err)
	}
}

func TestResolveTagSHAError(t *testing.T) {
	git := staticMockGit(nil, fmt.Errorf("not found"))

	_, err := git.ResolveTagSHA(context.Background(), "v1.0.0")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestFetchTags(t *testing.T) {
	git := staticMockGit([]byte(""), nil)

	if err := git.FetchTags(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestFetchTagsError(t *testing.T) {
	git := staticMockGit([]byte("fatal: could not read from remote repository\n"), fmt.Errorf("exit status 128"))

	err := git.FetchTags(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}
	for _, want := range []string{"failed to fetch tags", "exit status 128", "could not read from remote repository"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error %q does not contain %q", err, want)
		}
	}
}

func TestGetRemoteURL(t *testing.T) {
	git := staticMockGit([]byte("https://github.com/owner/repo.git\n"), nil)

	url, err := git.GetRemoteURL(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if url != "https://github.com/owner/repo.git" {
		t.Errorf("expected https://github.com/owner/repo.git, got %s", url)
	}
}

func TestGetRemoteURLError(t *testing.T) {
	git := staticMockGit(nil, fmt.Errorf("no remote"))

	_, err := git.GetRemoteURL(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestUpdateTag(t *testing.T) {
	tgr := newMockTagger(func(args ...string) ([]byte, error) {
		return []byte(""), nil
	})

	err := tgr.UpdateTag(context.Background(), "v1", "abc123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUpdateTagCreateError(t *testing.T) {
	tgr := newMockTagger(func(args ...string) ([]byte, error) {
		if args[0] == "tag" && len(args) > 2 {
			return nil, fmt.Errorf("create error")
		}
		return []byte(""), nil
	})

	err := tgr.UpdateTag(context.Background(), "v1", "abc123")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestUpdateTagPushError(t *testing.T) {
	tgr := newMockTagger(func(args ...string) ([]byte, error) {
		if args[0] == "push" {
			return nil, fmt.Errorf("push error")
		}
		return []byte(""), nil
	})

	err := tgr.UpdateTag(context.Background(), "v1", "abc123")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestRunSuccess(t *testing.T) {
	tgr := newMockTagger(func(args ...string) ([]byte, error) {
		if args[0] == "rev-list" {
			return []byte("abc123def456abc123def456abc123def456abc1\n"), nil
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

	err := tgr.ConfigureAuth(context.Background(), "mytoken", "")
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

	err := tgr.ConfigureAuth(context.Background(), "mytoken", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestConfigureAuthNoCredentials(t *testing.T) {
	tgr := newMockTagger(func(args ...string) ([]byte, error) {
		return []byte(""), nil
	})

	err := tgr.ConfigureAuth(context.Background(), "", "")
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

	err := tgr.ConfigureAuth(context.Background(), "mytoken", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestConfigureSSHAuth(t *testing.T) {
	isolateGitConfigEnv(t)
	home := t.TempDir()
	t.Setenv("HOME", home)

	var setURL string
	tgr := newMockTagger(func(args ...string) ([]byte, error) {
		if args[0] == "remote" && args[1] == "get-url" {
			return []byte("https://github.com/owner/repo.git\n"), nil
		}
		if args[0] == "remote" && args[1] == "set-url" {
			setURL = args[3]
		}
		return []byte(""), nil
	})

	err := tgr.ConfigureAuth(context.Background(), "", "fake-ssh-key-content")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if setURL != "git@github.com:owner/repo.git" {
		t.Errorf("origin set to %q, want git@github.com:owner/repo.git", setURL)
	}

	key, err := os.ReadFile(home + "/.ssh/id_rsa")
	if err != nil {
		t.Fatalf("failed to read key: %v", err)
	}
	if string(key) != "fake-ssh-key-content\n" {
		t.Errorf("key = %q, want a trailing newline appended", key)
	}

	knownHosts, err := os.ReadFile(home + "/.ssh/known_hosts")
	if err != nil {
		t.Fatalf("failed to read known_hosts: %v", err)
	}
	if string(knownHosts) != githubKnownHosts {
		t.Errorf("known_hosts = %q, want githubKnownHosts", knownHosts)
	}

	wantCmd := "core.sshCommand=ssh -i '" + home + "/.ssh/id_rsa' -o IdentitiesOnly=yes -o UserKnownHostsFile='" +
		home + "/.ssh/known_hosts' -o StrictHostKeyChecking=yes"
	if got := gitConfigEntries(t); !slices.Equal(got, []string{wantCmd}) {
		t.Errorf("git config env = %q, want [%q]", got, wantCmd)
	}
}

func TestConfigureSSHAuthKeepsTrailingNewline(t *testing.T) {
	isolateGitConfigEnv(t)
	home := t.TempDir()
	t.Setenv("HOME", home)

	tgr := NewTagger(staticMockGit(nil, nil))
	if err := tgr.ConfigureAuth(context.Background(), "", "fake-ssh-key-content\n"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	key, err := os.ReadFile(home + "/.ssh/id_rsa")
	if err != nil {
		t.Fatalf("failed to read key: %v", err)
	}
	if string(key) != "fake-ssh-key-content\n" {
		t.Errorf("key = %q, want it unchanged", key)
	}
}

func TestConfigureSSHAuthNonGitHubRemote(t *testing.T) {
	isolateGitConfigEnv(t)
	t.Setenv("HOME", t.TempDir())

	tgr := newMockTagger(func(args ...string) ([]byte, error) {
		if args[0] == "remote" && args[1] == "set-url" {
			t.Errorf("unexpected set-url for a non-GitHub origin: %v", args)
		}
		return []byte("https://gitlab.example.com/owner/repo.git\n"), nil
	})

	if err := tgr.ConfigureAuth(context.Background(), "", "fake-ssh-key"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestConfigureSSHAuthRemoteError(t *testing.T) {
	isolateGitConfigEnv(t)
	t.Setenv("HOME", t.TempDir())

	tgr := NewTagger(staticMockGit(nil, fmt.Errorf("no remote")))
	if err := tgr.ConfigureAuth(context.Background(), "", "fake-ssh-key"); err == nil {
		t.Fatal("expected error")
	}
}

func TestConfigureSSHAuthConfigEnvError(t *testing.T) {
	isolateGitConfigEnv(t)
	t.Setenv("HOME", t.TempDir())
	t.Setenv("GIT_CONFIG_COUNT", "bogus")

	tgr := NewTagger(staticMockGit(nil, nil))
	err := tgr.ConfigureAuth(context.Background(), "", "fake-ssh-key")
	if err == nil || !strings.Contains(err.Error(), "invalid GIT_CONFIG_COUNT") {
		t.Fatalf("expected invalid GIT_CONFIG_COUNT error, got: %v", err)
	}
}

// Fingerprints from docs.github.com "GitHub's SSH key fingerprints".
func TestGitHubKnownHostsFingerprints(t *testing.T) {
	want := map[string]string{
		"ssh-ed25519":         "SHA256:+DiY3wvvV6TuJJhbpZisF/zLDA0zPMSvHdkr4UvCOqU",
		"ecdsa-sha2-nistp256": "SHA256:p2QAMXNIC1TJYWeIOttrVc98/R1BUFWu3/LiyKgUfQM",
		"ssh-rsa":             "SHA256:uNiVztksCsDhcc0u9e8BujQXVUpKZIDTMczCvj3tD2s",
	}

	for _, line := range strings.Split(strings.TrimSuffix(githubKnownHosts, "\n"), "\n") {
		fields := strings.Fields(line)
		if len(fields) != 3 || fields[0] != "github.com" {
			t.Fatalf("malformed known_hosts line: %q", line)
		}
		keyType := fields[1]
		blob, err := base64.StdEncoding.DecodeString(fields[2])
		if err != nil {
			t.Fatalf("%s: invalid base64: %v", keyType, err)
		}
		sum := sha256.Sum256(blob)
		if got := "SHA256:" + base64.RawStdEncoding.EncodeToString(sum[:]); got != want[keyType] {
			t.Errorf("%s fingerprint = %s, want %s", keyType, got, want[keyType])
		}
		delete(want, keyType)
	}
	for keyType := range want {
		t.Errorf("missing %s host key", keyType)
	}
}

func TestConfigureTokenAuthRemoteError(t *testing.T) {
	tgr := newMockTagger(func(args ...string) ([]byte, error) {
		if args[0] == "remote" && args[1] == "get-url" {
			return nil, fmt.Errorf("no remote")
		}
		return []byte(""), nil
	})

	err := tgr.ConfigureAuth(context.Background(), "mytoken", "")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestRunUpdateMajorTagError(t *testing.T) {
	tgr := newMockTagger(func(args ...string) ([]byte, error) {
		if args[0] == "rev-list" {
			return []byte("abc123def456abc123def456abc123def456abc1\n"), nil
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
	// HOME is a regular file, so MkdirAll($HOME/.ssh) fails with ENOTDIR.
	tmpDir := t.TempDir()
	badPath := tmpDir + "/blocked"
	os.WriteFile(badPath, []byte("x"), 0444)
	t.Setenv("HOME", badPath)

	err := NewTagger(staticMockGit(nil, nil)).configureSSHAuth(context.Background(), "fake-key")
	if err == nil {
		t.Fatal("expected error for bad HOME path")
	}
}

func TestConfigureSSHAuthWriteKeyError(t *testing.T) {
	// .ssh dir exists but key path is a directory → WriteFile fails
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	os.MkdirAll(tmpDir+"/.ssh/id_rsa", 0700) // create dir where file should be

	err := NewTagger(staticMockGit(nil, nil)).configureSSHAuth(context.Background(), "fake-key")
	if err == nil {
		t.Fatal("expected error for write key failure")
	}
}

func TestConfigureSSHAuthWriteKnownHostsError(t *testing.T) {
	// key write succeeds, but known_hosts path is a directory → WriteFile fails
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	os.MkdirAll(tmpDir+"/.ssh/known_hosts", 0700) // create dir where file should be

	err := NewTagger(staticMockGit(nil, nil)).configureSSHAuth(context.Background(), "fake-key")
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

func TestExtractRepoPathURLForms(t *testing.T) {
	for _, remote := range []string{
		"ssh://git@github.com/owner/repo.git",
		"ssh://git@github.com:22/owner/repo.git",
		"https://user:pass@github.com/owner/repo.git",
		"https://github.com/owner/repo",
	} {
		if got := extractRepoPath(remote); got != "owner/repo" {
			t.Errorf("extractRepoPath(%q) = %q, want owner/repo", remote, got)
		}
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

func TestRunMinorTagUpdateError(t *testing.T) {
	callCount := 0
	tgr := newMockTagger(func(args ...string) ([]byte, error) {
		if args[0] == "rev-list" {
			return []byte("abc123def456abc123def456abc123def456abc1\n"), nil
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
	isolateGitConfigEnv(t)
	t.Setenv("GIT_CONFIG_COUNT", "bogus")
	tgr := newMockTagger(func(args ...string) ([]byte, error) {
		if args[0] == "rev-list" {
			return []byte("abc123def456abc123def456abc123def456abc1\n"), nil
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
		return []byte(""), nil
	})

	isolateGitConfigEnv(t)
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

	err := NewTagger(staticMockGit(nil, nil)).configureSSHAuth(context.Background(), "fake-key")
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
	mock := &MockRunner{Fn: func(args ...string) ([]byte, error) {
		return []byte("ok"), nil
	}}
	git := NewGit(mock)
	if git == nil {
		t.Fatal("expected non-nil Git")
	}

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

func TestCreateTag(t *testing.T) {
	git := staticMockGit([]byte(""), nil)
	if err := git.CreateTag(context.Background(), "v1", "abc123"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCreateTagError(t *testing.T) {
	git := staticMockGit(nil, fmt.Errorf("tag error"))
	err := git.CreateTag(context.Background(), "v1", "abc123")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestPushTag(t *testing.T) {
	git := staticMockGit([]byte(""), nil)
	if err := git.PushTag(context.Background(), "v1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestPushTagError(t *testing.T) {
	git := staticMockGit(nil, fmt.Errorf("push error"))
	err := git.PushTag(context.Background(), "v1")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestSetRemoteURL(t *testing.T) {
	git := staticMockGit([]byte(""), nil)
	if err := git.SetRemoteURL(context.Background(), "https://example.com"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSetRemoteURLError(t *testing.T) {
	git := staticMockGit(nil, fmt.Errorf("set-url error"))
	err := git.SetRemoteURL(context.Background(), "https://example.com")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestConfigureSafeDirectory(t *testing.T) {
	isolateGitConfigEnv(t)
	git := newMockGit(func(args ...string) ([]byte, error) {
		t.Errorf("expected no git call, got %v", args)
		return nil, nil
	})
	if err := git.ConfigureSafeDirectory(context.Background(), "/workspace"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []string{"safe.directory=/workspace"}
	if got := gitConfigEntries(t); !slices.Equal(got, want) {
		t.Errorf("expected entries %v, got %v", want, got)
	}
}

func TestConfigureSafeDirectoryError(t *testing.T) {
	isolateGitConfigEnv(t)
	t.Setenv("GIT_CONFIG_COUNT", "bogus")
	git := staticMockGit([]byte(""), nil)
	err := git.ConfigureSafeDirectory(context.Background(), "/workspace")
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

// TestUpdateTagNeverDeletesRemoteRef is the regression guard for the incident
// on 2026-08-07: UpdateTag deleted the remote tag, then failed on the push that
// would have recreated it, leaving `v1` absent. Absent is far worse than stale
// here — this action's own release workflow consumes itself as `@v1`, so a
// missing `v1` cannot be repaired by running the action.
func TestUpdateTagNeverDeletesRemoteRef(t *testing.T) {
	var calls [][]string
	tgr := newMockTagger(func(args ...string) ([]byte, error) {
		calls = append(calls, args)
		return []byte(""), nil
	})

	if err := tgr.UpdateTag(context.Background(), "v1", "abc123"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var pushed bool
	for _, c := range calls {
		joined := strings.Join(c, " ")
		if strings.Contains(joined, ":refs/tags/v1") && strings.HasPrefix(joined, "push origin :") {
			t.Fatalf("UpdateTag deleted the remote ref: %q", joined)
		}
		if c[0] == "tag" && len(c) > 1 && c[1] == "-d" {
			t.Fatalf("UpdateTag deleted the local ref: %q", joined)
		}
		if c[0] == "push" {
			pushed = true
			if !slices.Contains(c, "--force") {
				t.Errorf("push must be a force update, got %q", joined)
			}
			if !slices.Contains(c, "refs/tags/v1:refs/tags/v1") {
				t.Errorf("push must use a fully qualified refspec, got %q", joined)
			}
		}
	}
	if !pushed {
		t.Fatal("UpdateTag never pushed")
	}
}

// TestCreateTagMovesExistingTag pins the `-f`: without it, re-pointing a tag
// that already exists fails, which is the whole reason the delete came first.
func TestCreateTagMovesExistingTag(t *testing.T) {
	var got []string
	git := newMockGit(func(args ...string) ([]byte, error) {
		got = args
		return []byte(""), nil
	})

	if err := git.CreateTag(context.Background(), "v1", "abc123"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !slices.Contains(got, "-f") {
		t.Fatalf("CreateTag must force-move the tag, got %q", strings.Join(got, " "))
	}
}

// TestRunErrorIncludesGitOutput covers the second half of the same incident:
// the failure was reported as a bare "exit status 1", so why the push failed
// was unknowable from the run log.
func TestRunErrorIncludesGitOutput(t *testing.T) {
	git := staticMockGit([]byte("remote: Permission denied\nfatal: unable to access"), fmt.Errorf("exit status 1"))
	err := git.PushTag(context.Background(), "v1")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "Permission denied") {
		t.Fatalf("error must carry git's own output, got: %v", err)
	}
}

func TestRunErrorRedactsCredentials(t *testing.T) {
	git := staticMockGit([]byte("fatal: unable to access 'https://ghp_secrettoken@github.com/o/r/': 403"), fmt.Errorf("exit status 1"))
	err := git.PushTag(context.Background(), "v1")
	if err == nil {
		t.Fatal("expected error")
	}
	if strings.Contains(err.Error(), "ghp_secrettoken") {
		t.Fatalf("token leaked into error: %v", err)
	}
	if !strings.Contains(err.Error(), "https://***@github.com") {
		t.Fatalf("expected redacted remote, got: %v", err)
	}
}
