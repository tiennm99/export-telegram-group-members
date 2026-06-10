import sys

from compare import parse_args

RESET = '\033[0m'
YELLOW_BG = '\033[43m'
BLACK = '\033[30m'
DIVIDER = ' | '


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
    print_compare_ui(args.group_id, before_record, after_record, added, removed)
    return 0


def print_compare_ui(group_id, before_record, after_record, added, removed):
    title = after_record.get('title') or before_record.get('title') or ''
    before_members = before_record.get('members', [])
    after_members = after_record.get('members', [])
    before_time = before_record.get('time')
    after_time = after_record.get('time')
    label = f'{title} ({group_id})' if title else str(group_id)

    rows = side_by_side_rows(removed, added)
    widths = table_widths(rows)
    left_width = side_width(widths)
    right_width = left_width

    print(f'compare --telegram-group "{label}"')
    print(
        f'old: {before_time} ({len(before_members)} members)'.ljust(left_width)
        + DIVIDER
        + f'new: {after_time} ({len(after_members)} members)'.ljust(right_width)
    )
    print(
        f'removed: {len(removed)}'.ljust(left_width)
        + DIVIDER
        + f'added: {len(added)}'.ljust(right_width)
    )
    print()
    print(format_table_header(widths))
    print(format_separator(widths))

    for row in rows:
        print(highlight(format_table_row(row, widths)))

    if not added and not removed:
        print('no membership changes.')


def highlight(text):
    if not sys.stdout.isatty():
        return text
    return f'{YELLOW_BG}{BLACK}{text}{RESET}'


def side_by_side_rows(removed, added):
    rows = []
    max_count = max(len(removed), len(added))
    for index in range(max_count):
        left = member_columns(removed[index]) if index < len(removed) else empty_columns()
        right = member_columns(added[index]) if index < len(added) else empty_columns()
        rows.append((left, right))
    return rows


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


def empty_columns():
    return ['', '', '']


def table_widths(rows):
    widths = [2, 8, 4]
    for left, right in rows:
        for columns in (left, right):
            for index, value in enumerate(columns):
                widths[index] = max(widths[index], len(value))
    return widths


def side_width(widths):
    return sum(widths) + 4


def format_table_header(widths):
    labels = ['ID', 'Username', 'Name']
    left = format_columns(labels, widths)
    right = format_columns(labels, widths)
    return f'{left}{DIVIDER}{right}'


def format_separator(widths):
    return f'{"-" * side_width(widths)}-+-{"-" * side_width(widths)}'


def format_table_row(row, widths):
    left, right = row
    return f'{format_columns(left, widths)}{DIVIDER}{format_columns(right, widths)}'


def format_columns(columns, widths):
    return (
        f'{columns[0]:<{widths[0]}}  '
        f'{columns[1]:<{widths[1]}}  '
        f'{columns[2]:<{widths[2]}}'
    )


if __name__ == '__main__':
    raise SystemExit(main())
