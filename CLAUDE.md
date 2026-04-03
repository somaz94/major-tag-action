# CLAUDE.md

<br/>

<br/>

## Project Structure

- Go-based GitHub Action (Docker container action)
- Automatically updates major (and optionally minor) version tags to point to the latest semver release
- Example: releasing `v1.2.3` auto-updates `v1` tag to same commit

<br/>

## Build & Test

```bash
make build       # Build binary
make test        # Unit tests with coverage (85% threshold in CI)
make cover       # Generate coverage report
make fmt         # Format code
make lint        # Run go vet
```

<br/>

## Key Directories

- `cmd/` — Entry point (main.go, context/signal handling)
- `internal/config/` — Configuration from env vars (INPUT_*), with Validate() method
- `internal/tagger/` — Core logic with interface-based dependency injection:
  - `errors.go` — Sentinel errors (ErrInvalidTag, ErrAuthFailed, ErrTagUpdate, ErrInvalidSHA)
  - `git.go` — GitRunner interface, ExecRunner implementation, Git struct with methods
  - `tagger.go` — Tagger struct orchestrating tag update workflow
- `internal/output/` — GitHub Actions output helpers

<br/>

## Architecture

- **Dependency Injection**: `GitRunner` interface enables testability without package-level mocks
- **Construction**: `DefaultTagger()` → `NewTagger(NewGit(&ExecRunner{}))` for production
- **Testing**: `MockRunner` struct implements `GitRunner` for test isolation
- **Context**: `context.Context` propagated through `Tagger.Run()` for cancellation support

<br/>

## Action Inputs

Required: `tag` (full semver tag, e.g. `v1.2.3`)

Optional: `github_token`, `ssh_key`, `major_only` (default: true — set false to also update minor tag)

Outputs: `major_tag`, `minor_tag`, `commit_sha`

<br/>

## CI

- `ci.yml` — Unit tests (85% coverage threshold), Docker build, action validation
- Docker: multi-stage build (golang:1.26-alpine → alpine:3.23)

<br/>
