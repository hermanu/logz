# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Initial project scaffolding.
- `logz filter`, `logz summary`, `logz tail`, `logz fields` commands (skeleton).
- Pluggable parser interface with JSON and plain-text implementations.
- Configuration system (file + env + flags) using Koanf.
- CI on Linux, macOS, Windows; release pipeline via GoReleaser.
- Homebrew tap, Scoop bucket, Linux .deb/.rpm/.apk packages, multi-arch GHCR
  Docker images, and SBOM generation in the release pipeline.
- User-facing docs: `docs/install.md`, `docs/usage.md`, `docs/cookbook.md`.
- `AGENTS.md` and `CLAUDE.md` for AI coding agents.

### Changed
- `output.PrettyWriter` no longer mutates `fatih/color`'s package-level
  `NoColor` global; colors are disabled per-instance, fixing a data race when
  two writers ran concurrently.

### Tooling
- Tightened `.golangci.yml` to enable ~30 linters (added `paralleltest`,
  `tparallel`, `testpackage`, `thelper`, `funlen`, `gocyclo`, `nestif`,
  `nakedret`, `interfacebloat`, `forbidigo`, `gochecknoinits`, `goconst`,
  `errchkjson`, `gocheckcompilerdirectives`, `mirror`, `perfsprint`, `gci`).

[Unreleased]: https://github.com/hermanu/logz/compare/HEAD...HEAD
