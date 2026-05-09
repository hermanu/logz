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

[Unreleased]: https://github.com/hermanu/logz/compare/HEAD...HEAD
