package tagger

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

// isolateGitConfigEnv starts the test with no GIT_CONFIG_COUNT/KEY_<n>/VALUE_<n>
// set; t.Setenv restores the originals and the cleanup drops entries the test added.
func isolateGitConfigEnv(t *testing.T) {
	t.Helper()
	for _, name := range append(gitConfigEntryNames(), "GIT_CONFIG_COUNT") {
		t.Setenv(name, "")
		os.Unsetenv(name)
	}
	t.Cleanup(func() {
		for _, name := range gitConfigEntryNames() {
			os.Unsetenv(name)
		}
	})
}

func gitConfigEntryNames() []string {
	var names []string
	for _, kv := range os.Environ() {
		name, _, _ := strings.Cut(kv, "=")
		if strings.HasPrefix(name, "GIT_CONFIG_KEY_") || strings.HasPrefix(name, "GIT_CONFIG_VALUE_") {
			names = append(names, name)
		}
	}
	return names
}

// gitConfigEntries returns the injected entries as key=value, in index order.
func gitConfigEntries(t *testing.T) []string {
	t.Helper()
	n, err := strconv.Atoi(os.Getenv("GIT_CONFIG_COUNT"))
	if err != nil {
		t.Fatalf("GIT_CONFIG_COUNT: %v", err)
	}
	entries := make([]string, 0, n)
	for i := range n {
		entries = append(entries, os.Getenv(fmt.Sprintf("GIT_CONFIG_KEY_%d", i))+"="+os.Getenv(fmt.Sprintf("GIT_CONFIG_VALUE_%d", i)))
	}
	return entries
}

func TestAddConfigEnvFirstEntry(t *testing.T) {
	isolateGitConfigEnv(t)

	if err := AddConfigEnv("safe.directory", "/github/workspace"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := []string{"safe.directory=/github/workspace"}
	if got := gitConfigEntries(t); !reflect.DeepEqual(got, want) {
		t.Errorf("expected entries %v, got %v", want, got)
	}
}

func TestAddConfigEnvEmptyCount(t *testing.T) {
	isolateGitConfigEnv(t)
	t.Setenv("GIT_CONFIG_COUNT", "")

	if err := AddConfigEnv("safe.directory", "/github/workspace"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := []string{"safe.directory=/github/workspace"}
	if got := gitConfigEntries(t); !reflect.DeepEqual(got, want) {
		t.Errorf("expected entries %v, got %v", want, got)
	}
}

func TestAddConfigEnvAppendsToExisting(t *testing.T) {
	isolateGitConfigEnv(t)
	t.Setenv("GIT_CONFIG_COUNT", "2")
	t.Setenv("GIT_CONFIG_KEY_0", "safe.directory")
	t.Setenv("GIT_CONFIG_VALUE_0", "/runner/work")
	t.Setenv("GIT_CONFIG_KEY_1", "core.autocrlf")
	t.Setenv("GIT_CONFIG_VALUE_1", "false")

	if err := AddConfigEnv("safe.directory", "/github/workspace"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := []string{
		"safe.directory=/runner/work",
		"core.autocrlf=false",
		"safe.directory=/github/workspace",
	}
	if got := gitConfigEntries(t); !reflect.DeepEqual(got, want) {
		t.Errorf("expected entries %v, got %v", want, got)
	}
}

func TestAddConfigEnvInvalidCount(t *testing.T) {
	for _, count := range []string{"abc", "-1", "1.5"} {
		t.Run(count, func(t *testing.T) {
			isolateGitConfigEnv(t)
			t.Setenv("GIT_CONFIG_COUNT", count)

			if err := AddConfigEnv("safe.directory", "/github/workspace"); err == nil {
				t.Fatal("expected error for invalid GIT_CONFIG_COUNT")
			}
			if got := os.Getenv("GIT_CONFIG_COUNT"); got != count {
				t.Errorf("expected GIT_CONFIG_COUNT to stay %q, got %q", count, got)
			}
			if names := gitConfigEntryNames(); len(names) != 0 {
				t.Errorf("expected no entries written, got %v", names)
			}
		})
	}
}

func TestAddConfigEnvRejectsNUL(t *testing.T) {
	for name, kv := range map[string][2]string{
		"key":   {"safe.\x00directory", "/github/workspace"},
		"value": {"safe.directory", "/github/\x00workspace"},
	} {
		t.Run(name, func(t *testing.T) {
			isolateGitConfigEnv(t)

			if err := AddConfigEnv(kv[0], kv[1]); err == nil {
				t.Fatal("expected error for a NUL byte")
			}
			if _, ok := os.LookupEnv("GIT_CONFIG_COUNT"); ok {
				t.Error("expected GIT_CONFIG_COUNT to stay unset")
			}
		})
	}
}
