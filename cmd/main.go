package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/somaz94/major-tag-action/internal/config"
	"github.com/somaz94/major-tag-action/internal/output"
	"github.com/somaz94/major-tag-action/internal/tagger"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		output.LogWarning("Received shutdown signal, cleaning up...")
		cancel()
	}()
	defer func() {
		signal.Stop(sigCh)
		close(sigCh)
	}()

	if err := run(ctx); err != nil {
		output.LogError(err.Error())
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	cfg := config.Load()

	if cfg.Tag == "" {
		return fmt.Errorf("input 'tag' is required")
	}

	output.LogInfo("Starting major tag update...")

	select {
	case <-ctx.Done():
		return fmt.Errorf("cancelled")
	default:
	}

	result, err := tagger.Run(cfg.Tag, cfg.MajorOnly, cfg.GitHubToken, cfg.SSHKey)
	if err != nil {
		return fmt.Errorf("failed to update major tag: %w", err)
	}

	// Set outputs
	outputs := []struct {
		name  string
		value string
	}{
		{"major_tag", result.MajorTag},
		{"minor_tag", result.MinorTag},
		{"commit_sha", result.CommitSHA},
	}
	for _, o := range outputs {
		if err := output.SetOutput(o.name, o.value); err != nil {
			output.LogWarning(fmt.Sprintf("Failed to set %s output: %v", o.name, err))
		}
	}

	return nil
}
