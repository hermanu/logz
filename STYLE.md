# Style

This project follows the [Google Go Style Guide][google-go]. The rest of this
document is a checklist of the rules we actively enforce or care about — most
of it is covered by `golangci-lint` and `gofumpt`, but a few rules are about
intent, not syntax, and live here.

[google-go]: https://google.github.io/styleguide/go/

The three Google docs (in increasing specificity):

1. [Style Guide][google-go-guide] — the rules.
2. [Style Decisions][google-go-decisions] — the precise calls.
3. [Best Practices][google-go-best] — the patterns we copy.

[google-go-guide]: https://google.github.io/styleguide/go/guide
[google-go-decisions]: https://google.github.io/styleguide/go/decisions
[google-go-best]: https://google.github.io/styleguide/go/best-practices

## Naming

- **Acronyms keep their case**: `JSONParser`, `URL`, `ID`, `HTTP`. Not `Json`,
  `Url`, `Id`, `Http`.
- **No `Get` prefix on getters**: `Owner()`, not `GetOwner()`.
- **Receivers are short and consistent**: 1–2 chars, the same name on every
  method of the type. Never `self` or `this`.
- **Variable names scale with scope**: `i` is fine in a tight loop; `requestID`
  is required at function scope.
- **Package names are short, lowercase, and singular**: `parser`, not
  `parsers` or `parser_lib`.
- **Errors**: sentinel errors are `ErrFoo`; error types are `FooError`.
- **No type stutter**: `parser.Entry`, not `parser.ParserEntry`.

## Comments

- Every exported identifier has a doc comment.
- Doc comments start with the identifier name and are complete sentences
  ending with a period.
- Package docs go in a single file and start with `Package <name>`.
- Comments explain *why*, not *what*. The code shows what; the comment is for
  the parts the reader can't see.
- No commented-out code in commits. Use git history.

## Errors

- Wrap with `%w`. Never `%v` for an error. Never string concatenation.
- Compare with `errors.Is` / `errors.As`. Never `==` or string match.
- Error strings are lowercase, no trailing punctuation, no capital letters at
  the start of a sentence, and no newlines: `"could not connect to db"`, not
  `"Could not connect to DB."`
- Don't return both a non-nil result and a non-nil error unless the result is
  a partial that the caller must clean up.
- Sentinel errors live at package scope, prefixed `Err`, with a doc comment.

## Tests

- Table-driven where it pays off; subtests via `t.Run`.
- Every test that doesn't share state calls `t.Parallel()`.
- Helper functions call `t.Helper()`.
- Use [`go-cmp`][go-cmp] for structured comparisons; `t.Errorf` is fine for
  single values. Don't reach for `reflect.DeepEqual` or `testify`.
- Failure messages contain enough information to diagnose without re-running.
- Fuzz targets live alongside the parser they exercise. They assert "no panic
  and a reasonable Entry", not specific outputs.
- Examples (`ExampleFoo`) double as godoc and as compile-checked usage docs.
  Add them whenever an exported API has a non-obvious usage shape.

[go-cmp]: https://pkg.go.dev/github.com/google/go-cmp/cmp

## Concurrency

- Don't expose channels in public APIs. Wrap them in synchronous methods.
- Every goroutine must have a clear shutdown path. Use `context.Context` for
  cancellation; `select { case <-ctx.Done(): ... }` is the pattern.
- No global mutable state. (We had one slip — `fatih/color`'s `NoColor`
  package var — and fixed it. Don't reintroduce that pattern.)

## Layout

- `main.go` is a thin entrypoint.
- `cmd/` holds cobra commands and may import `internal/`.
- `internal/` packages depend only on each other and stdlib.
- No `init()` functions outside `main`. If you think you need one, you don't.

## Tooling

- `gofumpt` formats code (stricter than `gofmt`).
- `gci` orders imports: stdlib, third-party, project (separated by blank lines).
- `golangci-lint` — see [.golangci.yml](.golangci.yml). The current set of
  ~30 linters covers correctness, style, complexity, and test discipline.

Run `make check` before pushing. CI runs the same set on every PR and on
every push to `main`.
