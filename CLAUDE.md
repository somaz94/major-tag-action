# CLAUDE.md

<br/>

## Commit Guidelines

- Do not include `Co-Authored-By` lines in commit messages.
- Do not push to remote. Only commit. The user will push manually.

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
- `internal/config/` — Configuration from env vars (INPUT_*)
- `internal/tagger/` — Core logic: tagger.go (semver parsing, version extraction), git.go (tag operations)
- `internal/output/` — GitHub Actions output helpers

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

## Language

- Communicate with the user in Korean.
- All documentation and code comments must be written in English.
