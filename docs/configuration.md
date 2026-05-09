# Configuration

`logz` works without any configuration — sensible defaults handle most JSON
and plain-text logs. When your logs are unusual, configuration lets you teach
`logz` how to read them without writing code.

## Where config is loaded from

In order of precedence (later wins):

1. Built-in defaults
2. `$XDG_CONFIG_HOME/logz/config.yaml` (or `~/.config/logz/config.yaml`)
3. `./.logz.yaml` in the current directory
4. Environment variables prefixed `LOGZ_` (underscores become dots — e.g.
   `LOGZ_PARSER_LEVEL_KEYS=level,severity` → `parser.level_keys`)
5. CLI flags

A `--config <path>` flag bypasses the search and loads only the file you point
at (plus env + flags).

## Full example

```yaml
# ~/.config/logz/config.yaml

output: pretty       # pretty | json
no_color: false
top: 10              # default --top for `logz summary`

parser:
  # JSON / logfmt: which fields to read level/message/timestamp from.
  level_keys:     [level, severity, lvl]
  message_keys:   [msg, message, log]
  timestamp_keys: [ts, time, timestamp, "@timestamp"]

  # Extra time layouts to try in addition to RFC3339, common variants, etc.
  timestamp_layouts:
    - "2006-01-02 15:04:05.000 -0700"

  # Map non-standard level strings to canonical levels.
  # Useful for Bunyan (numeric) or syslog-style logs.
  level_aliases:
    "10": trace
    "20": debug
    "30": info
    "40": warn
    "50": error
    "60": fatal

  # For plain-text logs: a Go regexp with named capture groups.
  # Recognized group names: timestamp, level, message. Any other named group
  # becomes a structured field on the entry.
  text_pattern: '^(?P<timestamp>\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2})\s+\[(?P<level>[A-Z]+)\]\s+(?P<service>\S+)\s+(?P<message>.*)$'
```

## Tips

- Field names in `level_keys`, `message_keys`, `timestamp_keys` are tried in
  order; the first hit wins.
- `level_aliases` keys are matched case-insensitively after trimming.
- A custom `text_pattern` only applies when the parser falls through to text
  mode (auto-detect picked text, or you passed `--format text`).
- Unrecognized lines are kept and printed with the raw text as the message,
  so `--match` keyword filters still work even when the regex doesn't match.
