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
	if err := output.SetOutput("major_tag", result.MajorTag); err != nil {
		output.LogWarning(fmt.Sprintf("Failed to set major_tag output: %v", err))
	}
	if err := output.SetOutput("minor_tag", result.MinorTag); err != nil {
		output.LogWarning(fmt.Sprintf("Failed to set minor_tag output: %v", err))
	}
	if err := output.SetOutput("commit_sha", result.CommitSHA); err != nil {
		output.LogWarning(fmt.Sprintf("Failed to set commit_sha output: %v", err))
	}

	return nil
}
