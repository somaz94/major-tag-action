# Major Tag Action

[![Continuous Integration](https://github.com/somaz94/major-tag-action/actions/workflows/ci.yml/badge.svg)](https://github.com/somaz94/major-tag-action/actions/workflows/ci.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Latest Tag](https://img.shields.io/github/v/tag/somaz94/major-tag-action)](https://github.com/somaz94/major-tag-action/tags)
[![Top Language](https://img.shields.io/github/languages/top/somaz94/major-tag-action)](https://github.com/somaz94/major-tag-action)
[![GitHub Marketplace](https://img.shields.io/badge/Marketplace-Major%20Tag%20Action-blue?logo=github)](https://github.com/marketplace/actions/major-tag-action)

A Go-based GitHub Action that automatically updates major (and optionally minor) version tags to point to the latest semver release. For example, when you release `v1.2.3`, this action updates the `v1` tag to point to the same commit.

<br/>

## Features

- Automatically extract and update major version tags (e.g., `v1.2.3` -> `v1`)
- Optionally update minor version tags (e.g., `v1.2.3` -> `v1.2`)
- Support for both GitHub token and SSH key authentication
- Lightweight Go-based Docker container
- Outputs major tag, minor tag, and commit SHA

> For detailed usage examples, authentication options, and troubleshooting, see the [Usage Guide](docs/usage-guide.md).

## Usage

<br/>

### Basic

```yaml
steps:
  - name: Checkout
    uses: actions/checkout@v6
    with:
      fetch-depth: 0

  - name: Update major version tag
    uses: somaz94/major-tag-action@v1
    with:
      tag: ${{ github.ref_name }}
      github_token: ${{ secrets.PAT_TOKEN }}
```

<br/>

### In Release Workflow

```yaml
name: Create release

on:
  push:
    tags:
      - "v[0-9]+.[0-9]+.[0-9]+"

permissions:
  contents: write

jobs:
  release:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v6
        with:
          fetch-depth: 0

      - name: Create GitHub release
        uses: softprops/action-gh-release@v2
        with:
          tag_name: ${{ github.ref_name }}

      - name: Update major version tag
        uses: somaz94/major-tag-action@v1
        with:
          tag: ${{ github.ref_name }}
          github_token: ${{ secrets.PAT_TOKEN }}
```

<br/>

### With Minor Tag

```yaml
- name: Update major and minor version tags
  uses: somaz94/major-tag-action@v1
  with:
    tag: ${{ github.ref_name }}
    github_token: ${{ secrets.PAT_TOKEN }}
    major_only: false
```

This updates both `v1` and `v1.2` for a `v1.2.3` release.

<br/>

### With SSH Key

```yaml
- name: Update major version tag via SSH
  uses: somaz94/major-tag-action@v1
  with:
    tag: ${{ github.ref_name }}
    ssh_key: ${{ secrets.SSH_PRIVATE_KEY }}
```

<br/>

## Inputs

| Input | Description | Required | Default |
|-------|-------------|----------|---------|
| `tag` | The full semver tag (e.g., `v1.2.3`) | Yes | - |
| `github_token` | GitHub token for pushing tags | No | `${{ github.token }}` |
| `ssh_key` | SSH private key (alternative to token) | No | `` |
| `major_only` | Only update major tag. Set `false` to also update minor tag | No | `true` |

<br/>

## Outputs

| Output | Description | Example |
|--------|-------------|---------|
| `major_tag` | The major version tag updated | `v1` |
| `minor_tag` | The minor version tag updated (empty if `major_only: true`) | `v1.2` |
| `commit_sha` | The commit SHA the tags point to | `abc123def456` |

<br/>

## Why?

GitHub Actions users reference actions by major version (e.g., `uses: owner/action@v1`). When you release `v1.2.3`, the `v1` tag needs to be updated so users automatically get the latest patch. This action automates that process, replacing the common shell script pattern found in most release workflows.

<br/>

## Project Structure

```
.
├── cmd/
│   └── main.go              # Entry point
├── internal/
│   ├── config/
│   │   ├── config.go        # Configuration from env vars
│   │   └── config_test.go
│   ├── output/
│   │   ├── output.go        # GitHub Actions output helpers
│   │   └── output_test.go
│   └── tagger/
│       ├── git.go           # Git operations
│       ├── tagger.go        # Tag update logic
│       └── tagger_test.go
├── action.yml
├── Dockerfile
├── Makefile
└── go.mod
```

<br/>

## Development

<br/>

### Prerequisites

- Go 1.26+
- Docker (for container builds)

<br/>

### Build

```bash
make build
```

<br/>

### Test

```bash
make test
```

<br/>

### Coverage

```bash
make cover
```

<br/>

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

<br/>

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
