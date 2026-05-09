# Contributing to logz

Thanks for your interest in contributing! Issues, bug reports, and pull
requests are all welcome.

## Development setup

```bash
git clone https://github.com/hermanu/logz
cd logz
make tools   # installs golangci-lint, gofumpt, goreleaser locally
make test
```

Requires **Go 1.26+**.

## Workflow

1. Fork and create a feature branch from `main`.
2. Make your changes. Keep commits focused.
3. Run `make check` (vet + lint + test) before pushing.
4. Open a PR against `main`. Reference any related issue.

## Style

- `gofumpt` for formatting (stricter than `gofmt`).
- `golangci-lint` clean — see [.golangci.yml](.golangci.yml) for the ruleset.
- Add tests for new behavior. Aim for ≥80% coverage on `internal/`.
- No new dependencies without discussion. Stdlib first.

## Reporting bugs

Open a GitHub issue using the bug report template. Include `logz --version`,
a minimal log sample, and the command you ran.

## Security issues

See [SECURITY.md](SECURITY.md). Do not file public issues for vulnerabilities.

## Code of conduct

This project follows the [Contributor Covenant](CODE_OF_CONDUCT.md).
