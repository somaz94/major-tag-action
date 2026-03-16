package config

import (
	"testing"
)

func TestLoad(t *testing.T) {
	t.Setenv("INPUT_TAG", "v1.2.3")
	t.Setenv("INPUT_GITHUB_TOKEN", "ghp_test")
	t.Setenv("INPUT_SSH_KEY", "")
	t.Setenv("INPUT_MAJOR_ONLY", "true")

	cfg := Load()

	if cfg.Tag != "v1.2.3" {
		t.Errorf("expected v1.2.3, got %s", cfg.Tag)
	}
	if cfg.GitHubToken != "ghp_test" {
		t.Errorf("expected ghp_test, got %s", cfg.GitHubToken)
	}
	if cfg.SSHKey != "" {
		t.Errorf("expected empty SSH key, got %s", cfg.SSHKey)
	}
	if !cfg.MajorOnly {
		t.Error("expected MajorOnly to be true")
	}
}

func TestLoadMajorOnlyFalse(t *testing.T) {
	t.Setenv("INPUT_TAG", "v1.0.0")
	t.Setenv("INPUT_MAJOR_ONLY", "false")

	cfg := Load()

	if cfg.MajorOnly {
		t.Error("expected MajorOnly to be false")
	}
}

func TestLoadDefaults(t *testing.T) {
	t.Setenv("INPUT_TAG", "")
	t.Setenv("INPUT_GITHUB_TOKEN", "")
	t.Setenv("INPUT_SSH_KEY", "")
	t.Setenv("INPUT_MAJOR_ONLY", "")

	cfg := Load()

	if cfg.Tag != "" {
		t.Errorf("expected empty tag, got %s", cfg.Tag)
	}
	if !cfg.MajorOnly {
		t.Error("expected MajorOnly default to be true")
	}
}
