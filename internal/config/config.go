package config

import "os"

// Config holds all configuration for the major tag action.
type Config struct {
	Tag         string
	GitHubToken string
	SSHKey      string
	MajorOnly   bool
}

// Load reads configuration from environment variables (INPUT_*).
func Load() *Config {
	return &Config{
		Tag:         os.Getenv("INPUT_TAG"),
		GitHubToken: os.Getenv("INPUT_GITHUB_TOKEN"),
		SSHKey:      os.Getenv("INPUT_SSH_KEY"),
		MajorOnly:   getEnvDefault("INPUT_MAJOR_ONLY", "true") == "true",
	}
}

func getEnvDefault(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}
