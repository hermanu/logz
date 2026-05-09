# Usage

`logz` provides four subcommands:

| Command | Status | Purpose |
|---|---|---|
| `logz filter` | ✅ available | Filter and display log lines |
| `logz summary` | ✅ available | Summarize a log file |
| `logz tail` | 🚧 planned (v0.2) | Watch a log file in real time |
| `logz fields` | 🚧 planned (v0.2) | Inspect available fields |

Run `logz <command> --help` at any time for the full flag list.

## Global flags

These are accepted by every subcommand:

| Flag | Description |
|---|---|
| `--format <fmt>` | Force a parser: `json`, `text`. Default: auto-detect from the first non-empty line. |
| `--output, -o <mode>` | Output format: `pretty` (default) or `json` (NDJSON). |
| `--no-color` | Disable ANSI colors. Honored automatically when stdout isn't a TTY. |
| `--config <path>` | Use a specific config file instead of searching the standard locations. |
| `--version` | Print version, commit, and build date. |

If no file argument is provided, `logz` reads from stdin. This makes piping the
star feature:

```bash
kubectl logs my-pod | logz filter --level warn
journalctl -u nginx | logz summary
```

## `logz filter`

```
logz filter [flags] [file...]
```

| Flag | Description |
|---|---|
| `--level, -l <lvl>` | Minimum level: `trace`, `debug`, `info`, `warn`, `error`, `fatal`. |
| `--since <time>` | Drop entries before this time. Accepts durations (`1h`, `30m`) interpreted as relative to now, or absolute timestamps (RFC3339, `2024-01-01 12:00:00`, `2024-01-01`). |
| `--until <time>` | Drop entries after this time. Same syntax as `--since`. |
| `--match, -m <regex>` | Keep only entries where the line or message matches the Go regex. |
| `--field key=value` | Keep only entries where `Fields[key] == value`. Repeatable. |
| `--limit, -n <N>` | Stop after N matches. |
| `--invert, -v` | Invert the match (`grep -v` semantics). |

### Examples

```bash
# Errors in the last hour, mentioning "timeout"
logz filter --level error --since 1h --match timeout app.log

# All WARN+ from a specific service, as NDJSON for jq
logz filter --level warn --field service=auth -o json app.log | jq

# Last 100 lines that don't contain "healthcheck"
logz filter --invert --match healthcheck -n 100 app.log

# Combine multiple field filters
logz filter --field service=api --field method=POST app.log
```

### Exit codes

`logz filter` follows `grep`'s convention:

| Code | Meaning |
|---|---|
| 0 | At least one line matched |
| 1 | No matches |
| 2 | Error (bad flag, unreadable file, malformed regex, etc.) |

This makes it scriptable:

```bash
if ! logz filter --level fatal app.log >/dev/null; then
  echo "no fatals — happy day"
fi
```

## `logz summary`

```
logz summary [flags] [file...]
```

Aggregates a log file and prints a report:

- Total lines parsed, plus a count of unparseable lines.
- Time range (first and last timestamp).
- Lines per level.
- Average and peak events/sec.
- Top-N most frequent error messages.
- Top-N most common field names.

| Flag | Description |
|---|---|
| `--since <time>` | Only summarize entries after this time. |
| `--until <time>` | Only summarize entries before this time. |
| `--top N` | Number of rows in each top-N table. Defaults to the `top` value in your config (10). |

### Example

```bash
$ logz summary --top 5 app.log
Total lines:   12043
Time range:    2024-01-01T00:00:00Z → 2024-01-01T23:59:58Z

By level:
  info    11203
  warn      612
  error     219
  fatal       9

Events/sec:    avg 0.14, peak 47

Top 5 error messages:
    142  timeout connecting to db
     37  invalid auth token
     23  rate limit exceeded
     11  upstream 502
      6  panic: runtime error

Top 5 fields:
  12043  level
  12043  msg
  12043  ts
  10921  service
   8821  request_id
```

## Working with non-standard logs

If your logs use unusual field names, level encodings, or timestamp layouts,
configure `logz` rather than transforming the input. See
[configuration.md](configuration.md) for the full reference.

A common case: Bunyan-style numeric levels.

```yaml
# ~/.config/logz/config.yaml
parser:
  level_aliases:
    "10": trace
    "20": debug
    "30": info
    "40": warn
    "50": error
    "60": fatal
```

Then `logz filter --level warn` works as expected against logs whose `level`
field is the integer `40`.
