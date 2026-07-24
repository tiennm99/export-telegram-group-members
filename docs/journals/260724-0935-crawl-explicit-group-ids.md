# Crawl Now Requires Explicit Group IDs

---
date: 2026-07-24 09:35
session: crawl-explicit-group-ids
---

## Context

`crawl.py` previously crawled every `group_id` stored in Redis configuration.
The command now accepts its crawl targets explicitly so each run can select one
or more groups without reconfiguring Redis.

## What Happened

- Added `argparse` parsing to `crawl.py` with `group_ids` as a required positional argument.
- Changed `main()` to accept argv, parse IDs up front, and crawl only the supplied groups.
- Kept session reuse in Redis intact; this was a contract change for crawl targets, not auth or storage.
- Tests now cover multiple IDs, missing IDs, non-integer IDs, and the fact that configured `group_ids` are no longer used as an implicit fallback.

## Verification

`python -m unittest tests.test_crawl` passes. The test output shows the parser rejecting bad input as intended:

```text
python -m unittest: error: argument group_id: invalid int value: 'not-a-group'
python -m unittest: error: the following arguments are required: group_id
```

The happy-path test verifies `crawl.main(['101', '202'])` calls `get_entity(101)` and `get_entity(202)` and never touches the configured fallback `999`.

## Decisions

- Require at least one group ID instead of retaining a zero-argument fallback.
- Accept multiple positional IDs to support crawling several groups in one run.
- Keep stored `group_ids` unchanged for commands such as `compare.py` that use
  configured defaults.
- Parse arguments before loading Redis configuration or constructing a Telegram
  client, so invalid input fails without external contact.

## Next Steps

- Keep the README aligned with the new explicit-ID contract.
- Add coverage for any future CLI that mixes stored defaults with runtime arguments, because that pattern already proved easy to get wrong.
