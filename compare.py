import argparse
import sys


def parse_args():
    parser = argparse.ArgumentParser(
        description='Compare member changes between two crawls for one group.',
    )
    parser.add_argument('group_id', type=int, help='Telegram group id to compare')
    parser.add_argument(
        'times',
        nargs='*',
        help='Optional pair: time1 time2. When omitted, latest two crawls are used.',
    )
    args = parser.parse_args()
    if len(args.times) not in (0, 2):
        parser.error('provide both time1 and time2, or omit both to compare latest two crawls')
    return args


def main():
    args = parse_args()

    from common import diff_group_members, get_group_export, latest_two_group_exports

    if args.times:
        before_time, after_time = args.times
        before_record = get_group_export(args.group_id, before_time)
        after_record = get_group_export(args.group_id, after_time)
        if before_record is None:
            print(f'export not found for group {args.group_id} at {before_time}', file=sys.stderr)
            return 1
        if after_record is None:
            print(f'export not found for group {args.group_id} at {after_time}', file=sys.stderr)
            return 1
    else:
        before_record, after_record = latest_two_group_exports(args.group_id)
        if before_record is None or after_record is None:
            print(f'need at least 2 exports for group {args.group_id}', file=sys.stderr)
            return 1

    added, removed = diff_group_members(before_record, after_record)
    print_summary(args.group_id, before_record, after_record, added, removed)
    return 0


def print_summary(group_id, before_record, after_record, added, removed):
    title = after_record.get('title') or before_record.get('title') or ''
    before_members = before_record.get('members', [])
    after_members = after_record.get('members', [])

    print(f'group: {title} ({group_id})' if title else f'group: {group_id}')
    print(f'before: {before_record.get("time")} ({len(before_members)} members)')
    print(f'after:  {after_record.get("time")} ({len(after_members)} members)')
    print()

    print(f'added: {len(added)}')
    for member in added:
        print(f'+ {format_member(member)}')

    print()
    print(f'removed: {len(removed)}')
    for member in removed:
        print(f'- {format_member(member)}')

    if not added and not removed:
        print()
        print('no membership changes.')


def format_member(member):
    parts = [str(member.get('id'))]
    username = member.get('username')
    full_name = ' '.join(
        part for part in [member.get('first_name'), member.get('last_name')]
        if part
    )
    if username:
        parts.append(f'@{username}')
    if full_name:
        parts.append(full_name)
    return ' | '.join(parts)


if __name__ == '__main__':
    raise SystemExit(main())
