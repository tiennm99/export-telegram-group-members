import argparse
import sys

RESET = '\033[0m'
GREEN = '\033[32m'
RED = '\033[31m'
CYAN = '\033[36m'


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

    # Lazy import: keeps `python compare.py --help` working when REDIS_URL is unset.
    from common import diff_group_members, get_group_export, list_group_exports

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
        exports = list_group_exports(args.group_id)
        if len(exports) < 2:
            print(f'need at least 2 exports for group {args.group_id}', file=sys.stderr)
            return 1
        before_record, after_record = exports[-2:]

    added, removed = diff_group_members(before_record, after_record)
    print_summary(args.group_id, before_record, after_record, added, removed)
    return 0


def print_summary(group_id, before_record, after_record, added, removed):
    title = after_record.get('title') or before_record.get('title') or ''
    before_members = before_record.get('members', [])
    after_members = after_record.get('members', [])
    before_time = before_record.get('time')
    after_time = after_record.get('time')
    label = f'{title} ({group_id})' if title else str(group_id)

    print(f'diff --telegram-group "{label}"')
    print(color(f'--- crawl/{before_time}', RED))
    print(color(f'+++ crawl/{after_time}', GREEN))
    print(color(
        f'@@ members: {len(before_members)} -> {len(after_members)} '
        f'(+{len(added)} -{len(removed)}) @@',
        CYAN,
    ))
    print()

    widths = column_widths([*removed, *added])
    for member in removed:
        print(color(format_member_row('-', member, widths), RED))
    for member in added:
        print(color(format_member_row('+', member, widths), GREEN))

    if not added and not removed:
        print(' no membership changes.')


def color(text, code):
    if not sys.stdout.isatty():
        return text
    return f'{code}{text}{RESET}'


def member_columns(member):
    username = member.get('username')
    full_name = ' '.join(
        part for part in [member.get('first_name'), member.get('last_name')]
        if part
    )
    return [
        str(member.get('id')),
        f'@{username}' if username else '',
        full_name,
    ]


def column_widths(members):
    widths = [2, 8, 4]
    for member in members:
        for index, value in enumerate(member_columns(member)):
            widths[index] = max(widths[index], len(value))
    return widths


def format_member_row(prefix, member, widths):
    member_id, username, full_name = member_columns(member)
    return (
        f'{prefix} '
        f'{member_id:<{widths[0]}}  '
        f'{username:<{widths[1]}}  '
        f'{full_name:<{widths[2]}}'
    ).rstrip()


if __name__ == '__main__':
    raise SystemExit(main())
