# Task 18: GitHub Actions CI and Release Management

**Phase:** 5 — Infrastructure
**Blocked by:** 17
**Blocks:** none

## Objective

Set up GitHub Actions CI workflows and GoReleaser for automated releases with Homebrew tap publishing.

## Acceptance Criteria

### CI Workflow (`.github/workflows/ci.yml`)
- [ ] Triggered on push to main and pull requests
- [ ] Runs on ubuntu-latest
- [ ] Steps: checkout, setup Go (from go.mod), build, vet, test
- [ ] Go module caching enabled

### Release Workflow (`.github/workflows/release.yml`)
- [ ] Triggered on push of `v*` tags AND manual workflow_dispatch with version bump selection (major/minor/patch)
- [ ] Uses `goreleaser/goreleaser-action@v6`
- [ ] Passes `GITHUB_TOKEN` and `HOMEBREW_TAP_GITHUB_TOKEN` as env vars

### GoReleaser Config (`.goreleaser.yml`)
- [ ] Builds for linux/darwin, amd64/arm64
- [ ] CGO_ENABLED=0
- [ ] ldflags inject Version, Commit, Date into cmd package
- [ ] Archives as tar.gz (zip for windows if included)
- [ ] Checksums file
- [ ] Homebrew tap: `natikgadzhi/homebrew-taps` with Formula directory
- [ ] Changelog with filtered commit types

### Makefile Updates
- [ ] Update ldflags to include Commit and Date variables

## Notes

- Follow patterns from `natikgadzhi/gdrive-cli` closely
- Homebrew tap repo is `natikgadzhi/homebrew-taps`
- `HOMEBREW_TAP_GITHUB_TOKEN` is already set in repo secrets
- The release workflow should support both tag push and manual dispatch with version bump
