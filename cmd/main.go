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
	ctx, stop := notifyShutdown()
	defer stop()

	if err := run(ctx, tagger.DefaultTagger()); err != nil {
		output.LogError(err.Error())
		os.Exit(1)
	}
}

// notifyShutdown returns a context cancelled on SIGINT or SIGTERM, and a stop
// function that releases the handler. stop waits for the watcher to exit, so a
// normal exit never logs a shutdown.
func notifyShutdown() (context.Context, func()) {
	ctx, cancel := context.WithCancel(context.Background())
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	done := make(chan struct{})
	go func() {
		defer close(done)
		select {
		case <-sigCh:
			output.LogWarning("Received shutdown signal, cleaning up...")
			cancel()
		case <-ctx.Done():
		}
	}()

	return ctx, func() {
		signal.Stop(sigCh)
		cancel()
		<-done
	}
}

func run(ctx context.Context, t *tagger.Tagger) error {
	cfg := config.Load()

	if err := cfg.Validate(); err != nil {
		return err
	}

	output.LogInfo("Starting major tag update...")

	select {
	case <-ctx.Done():
		return fmt.Errorf("cancelled")
	default:
	}

	result, err := t.Run(ctx, cfg.Tag, cfg.MajorOnly, cfg.GitHubToken, cfg.SSHKey)
	if err != nil {
		return fmt.Errorf("failed to update major tag: %w", err)
	}

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
			output.LogWarning("Failed to set " + o.name + " output: " + err.Error())
		}
	}

	return nil
}
