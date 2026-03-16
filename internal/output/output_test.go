package output

import (
	"os"
	"strings"
	"testing"
)

func TestSetOutputToFile(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "github_output")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	t.Setenv("GITHUB_OUTPUT", tmpFile.Name())

	err = SetOutput("test_key", "test_value")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(tmpFile.Name())
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(string(data), "test_key=test_value") {
		t.Errorf("expected output to contain 'test_key=test_value', got: %s", string(data))
	}
}

func TestSetOutputMultiline(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "github_output")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	t.Setenv("GITHUB_OUTPUT", tmpFile.Name())

	err = SetOutput("content", "line1\nline2")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(tmpFile.Name())
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(string(data), "content<<EOF") {
		t.Errorf("expected multiline delimiter, got: %s", string(data))
	}
}

func TestSetOutputFallback(t *testing.T) {
	t.Setenv("GITHUB_OUTPUT", "")

	err := SetOutput("key", "value")
	if err != nil {
		t.Fatalf("unexpected error in fallback mode: %v", err)
	}
}

func TestSetOutputOpenFileError(t *testing.T) {
	t.Setenv("GITHUB_OUTPUT", "/nonexistent/dir/output")

	err := SetOutput("key", "value")
	if err == nil {
		t.Fatal("expected error when GITHUB_OUTPUT path is invalid")
	}
}

func TestLogInfo(t *testing.T) {
	r, w, _ := os.Pipe()
	oldStdout := os.Stdout
	os.Stdout = w

	LogInfo("test message")

	w.Close()
	os.Stdout = oldStdout

	buf := make([]byte, 256)
	n, _ := r.Read(buf)
	out := string(buf[:n])

	if !strings.Contains(out, "::notice::test message") {
		t.Errorf("expected '::notice::test message', got %q", out)
	}
}

func TestLogWarning(t *testing.T) {
	r, w, _ := os.Pipe()
	oldStdout := os.Stdout
	os.Stdout = w

	LogWarning("warn msg")

	w.Close()
	os.Stdout = oldStdout

	buf := make([]byte, 256)
	n, _ := r.Read(buf)
	out := string(buf[:n])

	if !strings.Contains(out, "::warning::warn msg") {
		t.Errorf("expected '::warning::warn msg', got %q", out)
	}
}

func TestLogError(t *testing.T) {
	r, w, _ := os.Pipe()
	oldStdout := os.Stdout
	os.Stdout = w

	LogError("error msg")

	w.Close()
	os.Stdout = oldStdout

	buf := make([]byte, 256)
	n, _ := r.Read(buf)
	out := string(buf[:n])

	if !strings.Contains(out, "::error::error msg") {
		t.Errorf("expected '::error::error msg', got %q", out)
	}
}
