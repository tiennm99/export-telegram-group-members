# Compare Group Export Diffs

## Context

- `common.py` stores and lists Redis export records.
- `crawl.py` exports groups and should not be reused for comparison because it logs into Telegram.
- `README.md` documents current usage and Redis record shape.

## Requirements

- Add command that receives `group_id`, optional `time1`, optional `time2`.
- If both times missing, compare latest 2 exports for that group.
- If one time missing, fail with clear usage error.
- Print members added and removed between two crawls.
- Keep Redis schema unchanged.

## Implementation

1. Add focused helpers in `common.py` to filter exports by group/time and compare members by Telegram `id`.
2. Add `compare.py` CLI with argparse.
3. Update `README.md` with command examples.
4. Run Python compile check.

## Success Criteria

- `python compare.py <group_id>` compares latest two times.
- `python compare.py <group_id> <time1> <time2>` compares explicit times.
- Missing export records produce readable errors.
- Existing export flow remains unchanged.

## Unresolved Questions

None.
