# Usage Guide

This guide explains how to use **major-tag-action** in your GitHub Actions workflows.

<br/>

## Table of Contents

- [How It Works](#how-it-works)
- [Prerequisites](#prerequisites)
- [Basic Usage](#basic-usage)
- [Release Workflow Example](#release-workflow-example)
- [Advanced Usage](#advanced-usage)
  - [Update Minor Tag](#update-minor-tag)
  - [SSH Key Authentication](#ssh-key-authentication)
  - [Using Outputs](#using-outputs)
- [Authentication](#authentication)
- [Tag Format](#tag-format)
- [Troubleshooting](#troubleshooting)

<br/>

## How It Works

When you push a semver tag (e.g., `v1.2.3`), this action:

1. Extracts the major version (`v1`) from the tag
2. Deletes the existing `v1` tag (if it exists)
3. Creates a new `v1` tag pointing to the same commit as `v1.2.3`
4. Pushes the updated tag to the remote

This ensures that users referencing `@v1` always get the latest release.

```
v1.0.0 ──> v1.1.0 ──> v1.2.3 (latest)
                          │
                          ▼
                         v1 (always points to latest v1.x.x)
```

<br/>

## Prerequisites

- Repository must have at least one semver tag (e.g., `v1.0.0`)
- `fetch-depth: 0` is required in the checkout step to access all tags
- A token with push permissions (`contents: write`)

<br/>

## Basic Usage

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

> **Note:** `github.token` (default) may not have sufficient permissions to push tags in some repository configurations. Use a Personal Access Token (PAT) with `contents: write` scope if needed.

<br/>

## Release Workflow Example

A complete release workflow that creates a GitHub release and updates the major version tag:

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
      - name: Checkout
        uses: actions/checkout@v6
        with:
          fetch-depth: 0
          token: ${{ secrets.PAT_TOKEN }}

      - name: Create GitHub release
        uses: softprops/action-gh-release@v2
        with:
          tag_name: ${{ github.ref_name }}
          name: ${{ github.ref_name }}
          generate_release_notes: true

      - name: Update major version tag
        uses: somaz94/major-tag-action@v1
        with:
          tag: ${{ github.ref_name }}
          github_token: ${{ secrets.PAT_TOKEN }}
```

<br/>

## Advanced Usage

### Update Minor Tag

Set `major_only: false` to also update the minor version tag:

```yaml
- name: Update major and minor version tags
  uses: somaz94/major-tag-action@v1
  with:
    tag: v1.2.3
    github_token: ${{ secrets.PAT_TOKEN }}
    major_only: false
```

This creates/updates both:
- `v1` -> points to `v1.2.3`
- `v1.2` -> points to `v1.2.3`

<br/>

### SSH Key Authentication

Use an SSH deploy key instead of a GitHub token:

```yaml
- name: Update major version tag
  uses: somaz94/major-tag-action@v1
  with:
    tag: ${{ github.ref_name }}
    ssh_key: ${{ secrets.SSH_PRIVATE_KEY }}
```

> **Note:** The SSH key must have write access to the repository.

<br/>

### Using Outputs

Access the action's outputs in subsequent steps:

```yaml
- name: Update major tag
  id: update-tag
  uses: somaz94/major-tag-action@v1
  with:
    tag: ${{ github.ref_name }}
    github_token: ${{ secrets.PAT_TOKEN }}
    major_only: false

- name: Print results
  run: |
    echo "Major tag: ${{ steps.update-tag.outputs.major_tag }}"
    echo "Minor tag: ${{ steps.update-tag.outputs.minor_tag }}"
    echo "Commit SHA: ${{ steps.update-tag.outputs.commit_sha }}"
```

<br/>

## Authentication

| Method | Input | When to use |
|--------|-------|-------------|
| GitHub Token (default) | `github_token` | Most common. Use `${{ github.token }}` or a PAT |
| SSH Key | `ssh_key` | When token-based auth is restricted or for deploy keys |

**Priority:** If both `ssh_key` and `github_token` are provided, SSH key takes precedence.

<br/>

## Tag Format

The action expects semver tags in the format `vX.Y.Z`:

| Input Tag | Major Tag | Minor Tag |
|-----------|-----------|-----------|
| `v1.0.0` | `v1` | `v1.0` |
| `v1.2.3` | `v1` | `v1.2` |
| `v2.0.0-rc1` | `v2` | `v2.0` |
| `v12.34.56` | `v12` | `v12.34` |

Tags that don't match the `vX.Y.Z` format will cause an error.

<br/>

## Troubleshooting

### "failed to push tag" error

- Ensure the token has `contents: write` permission
- Check that the repository allows tag force-push
- If using `github.token`, try a PAT instead

### "failed to resolve SHA" error

- Make sure `fetch-depth: 0` is set in the checkout step
- The tag must exist in the repository before running this action

### "does not match semver format" error

- The `tag` input must match `vX.Y.Z` format (e.g., `v1.0.0`)
- Tags without the `v` prefix are not supported

### Tag not updating for users

- Users must use `@v1` (not `@v1.0.0`) to get automatic updates
- Clear GitHub Actions cache if the old version is cached
