# Journal: Default Compare Group

---
date: 2026-07-15
session: default-compare-group
---

## Context

`compare.py` required a group ID even though the stored configuration already has an ordered `group_ids` list. The change makes the common latest-two-crawls comparison available with no positional arguments.

## What Happened

- Made `group_id` optional and resolved an omission to the first configured group.
- Kept explicit group and explicit group-plus-two-timestamps behavior unchanged.
- Preserved lazy imports so `python compare.py --help` works without `REDIS_URL`.
- Documented the zero-argument command in `README.md`.
- Independent review passed. Isolated checks covered zero arguments/default first group, explicit group, explicit group with timestamps, invalid arity, help without Redis, syntax, and diff integrity. The repository has no automated tests.

## Reflection

The implementation stayed small by reusing `load_app_config()` and normalizing the group ID once before existing lookup logic. CLI validation remains clear. The positional grammar cannot express explicit timestamps while omitting the group; that limitation is outside the documented scope.

## Decisions

| Decision | Rationale | Impact |
|---|---|---|
| Default to the first configured group | Configuration order already defines a practical default | `python compare.py` compares its latest two crawls |
| Keep timestamps coupled to an explicit group | Avoid ambiguous positional parsing and scope expansion | Timestamp selection remains `group_id time1 time2` |
| Retain lazy Redis-dependent imports | Help should not require runtime configuration | `--help` remains available without `REDIS_URL` |

## Next

- Add automated CLI tests if a test harness is introduced.
- Revisit named timestamp options only if timestamp comparison for the default group becomes a requirement.
