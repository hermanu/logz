# logz

[![CI](https://github.com/hermanu/logz/actions/workflows/ci.yml/badge.svg)](https://github.com/hermanu/logz/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/hermanu/logz)](https://goreportcard.com/report/github.com/hermanu/logz)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![Go Reference](https://pkg.go.dev/badge/github.com/hermanu/logz.svg)](https://pkg.go.dev/github.com/hermanu/logz)

A fast, developer-friendly CLI for parsing, filtering, and summarizing log files.

`logz` answers the questions you actually have when staring at a log file:

- *How many errors happened in the last hour?*
- *Show me all WARN+ from service X.*
- *What are the top 10 most frequent error messages?*

It auto-detects common log formats (JSON, logfmt, plain text, Apache/Nginx),
streams files line-by-line (multi-GB safe), and ships as a single static binary.

## Install

```bash
go install github.com/hermanu/logz@latest
```

Pre-built binaries for Linux, macOS, and Windows are attached to each
[release](https://github.com/hermanu/logz/releases).

## Quick start

```bash
# Filter
logz filter --level error --since 1h --match "timeout" app.log

# Summarize
logz summary --top 5 app.log

# Tail with filter
logz tail --level error app.log

# Pipe from kubectl, etc.
kubectl logs my-pod | logz filter --level warn
```

## Configuration

`logz` reads optional config from (in order of precedence, highest first):

1. CLI flags
2. Environment variables (`LOGZ_*`)
3. `./.logz.yaml` (project-local)
4. `$XDG_CONFIG_HOME/logz/config.yaml` (user)
5. Built-in defaults

Use config to teach `logz` about your logs — custom field names, level
mappings, timestamp layouts, color themes, and named text-format regexes.
See [docs/configuration.md](docs/configuration.md) for the full reference.

## Documentation

- [Configuration reference](docs/configuration.md)
- [Contributing](CONTRIBUTING.md)
- [Security policy](SECURITY.md)
- [Changelog](CHANGELOG.md)

## License

MIT — see [LICENSE](LICENSE).
