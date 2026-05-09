# Cookbook

Real-world recipes. Most assume your logs are JSON; the same patterns work for
plain-text logs but you'll lose field-level filtering.

## Find every error from the last 15 minutes

```bash
logz filter --level error --since 15m app.log
```

## How many distinct error messages are there?

```bash
logz summary --since 1h app.log | grep -A 100 'Top'
```

Or for a focused count:

```bash
logz filter --level error --since 1h -o json app.log \
  | jq -r '.message' \
  | sort -u \
  | wc -l
```

## Stream from `kubectl` and filter

```bash
kubectl logs -f deploy/api | logz filter --level warn --no-color
```

`logz` reads from stdin when no file is given. `--no-color` is unnecessary if
the output is going to your terminal, but useful when piping further.

## Filter by request ID across multiple files

```bash
logz filter --field request_id=abc123 *.log
```

## Strip noise — only POST requests that took >500ms

This requires a `duration_ms` field. The `--field` filter does exact match, so
for ranges drop into NDJSON + jq:

```bash
logz filter --field method=POST -o json app.log \
  | jq 'select(.fields.duration_ms | tonumber > 500)'
```

## Count requests per second

```bash
logz summary app.log | grep "Events/sec"
```

For a time-series breakdown, pipe to NDJSON and bucket with `awk` or `datamash`:

```bash
logz filter -o json app.log \
  | jq -r '.timestamp' \
  | cut -c1-19 \
  | sort | uniq -c
```

## Tail-and-filter (until v0.2 lands)

Until `logz tail` is implemented, `tail -F | logz filter` works for the common
case (no rotation handling):

```bash
tail -F app.log | logz filter --level warn
```

## Compare two days side by side

```bash
diff \
  <(logz summary --since "2024-01-01" --until "2024-01-02" app.log) \
  <(logz summary --since "2024-01-02" --until "2024-01-03" app.log)
```

## CI: fail the build on fatal errors in test logs

```bash
go test ./... 2>&1 | tee test.log
if logz filter --level fatal -n 1 test.log >/dev/null; then
  echo "::error::Fatal log line detected"
  exit 1
fi
```

(`-n 1` stops after the first match for speed; the exit code is 0 if any
matched, which is what triggers the failure.)

## Custom plain-text format

For logs that don't match the default text pattern, supply your own regex with
named capture groups in `~/.config/logz/config.yaml`:

```yaml
parser:
  text_pattern: '^(?P<timestamp>\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}) \[(?P<level>[A-Z]+)\] (?P<service>\S+) (?P<message>.*)$'
```

Recognized group names are `timestamp`, `level`, `message`. Any other named
group becomes a structured field on the entry — searchable via `--field`.

## NDJSON pipeline patterns

`logz filter -o json` produces one JSON object per line. Combine with `jq`:

```bash
# Average duration of error-tagged requests
logz filter --level error -o json app.log \
  | jq -s 'map(.fields.duration_ms | tonumber) | add / length'

# Top 10 user_ids by error count
logz filter --level error -o json app.log \
  | jq -r '.fields.user_id' \
  | sort | uniq -c | sort -rn | head
```
