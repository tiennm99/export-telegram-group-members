# Redis-to-CSV Export

---
date: 2026-07-24 10:25
session: redis-csv-export
---

## Context

Saved Telegram crawls lived only in Redis. The new command makes one crawl
available as a local CSV without contacting Telegram.

## What Happened

- Added `export.py [group_id] [timecrawl]`.
- Used the latest Redis-key timestamp when `timecrawl` is omitted.
- Used the first configured group when both arguments are omitted.
- Wrote `id`, `username`, `first_name`, and `last_name` to
  `output/<group-id>-<yyyymmddhhmmss>.csv`.
- Added `output/` to `.gitignore` and documented the command.

## Decisions

- Validate CLI timestamps, Redis record metadata, and every member before
  creating output.
- Fail on a corrupt newest crawl instead of silently exporting stale history.
- Neutralize spreadsheet-formula prefixes in Telegram text.
- Write through a temporary file and atomically replace the destination.
- Use `0700` directory and `0600` file permissions on POSIX systems.

## Verification

- All 18 unit tests passed.
- Python compilation and whitespace checks passed.
- CLI help works without `REDIS_URL`.
- Independent testing and review found no remaining issues.

## Next Steps

- Keep CSV validation and README documentation aligned if export fields change.
