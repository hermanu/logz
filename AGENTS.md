# AGENTS.md

Guidance for AI coding agents (Claude Code, Cursor, Copilot Workspace, Aider,
etc.) working on this repository. Humans should read [CONTRIBUTING.md](CONTRIBUTING.md);
this file captures the conventions and expectations specifically useful to an
agent making changes.

## Project overview

`logz` is a Go CLI for parsing, filtering, and summarizing log files. It ships
as a single static binary, supports JSON and plain-text formats out of the box
(logfmt and Apache/Nginx are planned), and is heavily configurable so users can
adapt it to non-standard logs without writing code.

The MVP scope and roadmap live in the project PRD; the relevant near-term goals
are tracked in `CHANGELOG.md` under `[Unreleased]`.

## Layout

```
.
├── main.go                      # entry point — defers to cmd.Run
├── cmd/                         # cobra commands; one file per subcommand
│   ├── root.go                  # root command + global flags
│   ├── pipeline.go              # shared file-open / parse / output plumbing
│   ├── filter.go  summary.go    # implemented commands
│   ├── tail.go    fields.go     # stubbed; planned for v0.2
│   └── cmd_test.go              # end-to-end CLI tests through cobra
├── internal/
│   ├── parser/                  # Entry, Level, Parser interface, JSON, text
│   ├── filter/                  # predicate engine (level, time, regex, fields)
│   ├── output/                  # pretty + NDJSON writers
│   ├── summary/                 # aggregation + report renderer
│   ├── config/                  # koanf-based config loader
│   └── logz/                    # version metadata (set via -ldflags)
├── docs/configuration.md        # user-facing config reference
└── testdata/                    # sample fixtures used by tests
```

The `cmd/` package is allowed to import `internal/`. Packages under
`internal/` should depend only on each other (and stdlib), never on `cmd/`.

## Toolchain

- **Go 1.26+** is required (see `go.mod`).
- Dev tools live in `./bin` after `make tools`: `golangci-lint`, `gofumpt`, `goreleaser`.
- CI runs on Linux, macOS, Windows with Go 1.26.

## Common commands

```bash
make build      # produce ./logz
make test       # go test -race ./...
make cover      # coverage.html
make lint       # golangci-lint run ./...
make check      # vet + lint + test (run this before pushing)
make fmt        # gofumpt -l -w .
make snapshot   # local goreleaser build (no publish)
```

For a quick sanity check after a change:

```bash
make check
```

## Conventions

### Errors

- Wrap with `%w` whenever propagating; never `%v` for an error value.
- Use `errors.Is` / `errors.As` for comparisons; never `==` or string matching.
- `parser.ErrSkip` is the sentinel for "this line should be silently dropped"
  — don't reuse it for real errors.
- Validate at boundaries (CLI flags, file paths, config files). Trust internal
  callers; don't re-validate inputs that internal code already produced.

### Style

- `gofumpt`-formatted (stricter than `gofmt`).
- `golangci-lint` clean — see `.golangci.yml`.
- Prefer `any` over `interface{}`.
- Comments on exported identifiers, leading with the identifier name (Go convention).
- No comments that just restate what the code does. Save comments for *why*
  something non-obvious was chosen.
- No emoji in source or markdown unless the user explicitly asks.

### Tests

- Table-driven tests where it helps; `t.Parallel()` on every test that doesn't
  share state.
- Test fixtures live in `testdata/` (compiler-recognized, ignored by `go vet`).
- New packages should ship with tests in the same commit, not as a follow-up.
- Aim for ≥80% coverage on `internal/`. Don't game the metric — coverage on
  trivial getters is meaningless; coverage on parsing edge cases matters.
- End-to-end CLI tests go in `cmd/cmd_test.go` and exercise commands through
  `cmd.Run` with `bytes.Buffer` stdin/stdout.

### Adding a new parser

1. Implement `parser.Parser` (Name + Parse) in `internal/parser/<format>.go`.
2. Update `parser.Detect` to recognize the format from a sample.
3. Add `<format>_test.go` with happy-path, edge-case, and malformed-input tests.
4. Add a fixture under `testdata/<format>/sample.log`.
5. Wire the format name into `cmd.pickParser` so `--format <name>` works.
6. Update `README.md` and `docs/configuration.md` if the parser introduces
   new config knobs.

### Adding a new command

1. Create `cmd/<command>.go` with a `new<Command>Cmd()` constructor.
2. Register it in `cmd.New` (`root.go`).
3. Use the helpers in `cmd/pipeline.go` (`resolveSources`, `pickParser`,
   `pickOutput`, `streamLines`) — don't reinvent the file/parser/output wiring.
4. Add a CLI test in `cmd/cmd_test.go`.

### Adding a config knob

1. Extend the `Config` (or `ParserConfig`) struct in `internal/config/config.go`
   with a `koanf:"snake_case"` tag.
2. Set a sensible default in `config.Default()`.
3. If the value flows into the parser, surface it through `Config.ParserOptions()`.
4. Document it in `docs/configuration.md` with an example.
5. Add a test in `internal/config/config_test.go` that loads a YAML file and
   asserts the value is propagated.

### Dependencies

The PRD constrains dependencies tightly. Current allowed deps:
`github.com/spf13/cobra`, `github.com/fatih/color`, `github.com/knadh/koanf/*`,
`github.com/fsnotify/fsnotify` (when tail lands), `github.com/go-logfmt/logfmt`
(when logfmt lands).

Do **not** add new direct dependencies without explicit user approval. If you
think one is necessary, raise it in the PR description rather than adding it.

## Things to avoid

- Don't load whole files into memory — everything is streaming. `bufio.Reader`
  with a 10MB ceiling, not `io.ReadAll`.
- Don't add concurrency unless the PRD calls for it (multi-file in v1.1).
  Single-goroutine pipelines are easier to reason about and the bottleneck is
  almost always I/O.
- Don't change the public CLI surface (flag names, exit codes) without
  flagging it in `CHANGELOG.md` as a breaking change.
- Don't break the `[Unreleased]` section of `CHANGELOG.md` — add entries there
  for any user-visible change.

## Commits and PRs

- Conventional-commit-ish prefixes (`feat:`, `fix:`, `refactor:`, `docs:`,
  `test:`, `chore:`) — used by GoReleaser to group release notes.
- One logical change per PR; don't bundle unrelated cleanup with feature work.
- Run `make check` locally before pushing.
- The PR template asks you to confirm `make check` passes and that
  `CHANGELOG.md` was updated; please actually do both.
