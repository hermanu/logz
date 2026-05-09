# CLAUDE.md

Project-specific instructions for [Claude Code](https://claude.com/claude-code).

The full set of conventions for AI agents working on this repo lives in
[AGENTS.md](AGENTS.md) — read it before making changes. CLAUDE.md only adds
Claude-specific notes on top.

## Scope

- This is an open-source Go CLI. Quality matters more than speed; prefer doing
  the right thing in a focused PR over patching things in.
- Don't introduce new direct dependencies without asking first. The PRD keeps
  the dep list intentionally short.
- Don't generate planning, decision, or analysis documents in the repo unless
  explicitly asked.

## Before you finish a task

1. Run `make check` (vet + lint + test). It must pass.
2. If user-visible behavior changed, add a `CHANGELOG.md` entry under
   `[Unreleased]`.
3. If a public API in `internal/` changed, update its godoc on the exported
   identifier — `golangci-lint`'s `revive` rule will flag missing docs anyway.
4. Confirm cross-platform safety: no hard-coded `/` separators in user-facing
   paths, no shell-only constructs in test commands. Tests must pass on
   Windows too.

## Avoid

- Don't ship code that compiles only on `main` HEAD — anything that lands
  needs to compile and pass tests on the matrix in `.github/workflows/ci.yml`.
- Don't replace the koanf-based config system without discussion. Configurability
  is a deliberate product goal, not an accident.
- Don't promote `tail` / `fields` from stub to real implementation as a
  drive-by — they're scoped for v0.2 and warrant focused PRs (fsnotify
  semantics for `tail` need their own review).
- Don't widen the CLI surface (new commands or flags) without first
  discussing with the user.
