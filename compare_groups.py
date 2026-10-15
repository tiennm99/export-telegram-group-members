import argparse
import sys

from compare import CYAN, GREEN, RED, color, column_widths, format_member_row


def parse_args(argv=None):
    parser = argparse.ArgumentParser(
        description=(
            'Fetch two Telegram groups live and compare their members. '
            'Nothing is saved to Redis.'
        ),
    )
    parser.add_argument('group_a', type=int, help='first Telegram group id')
    parser.add_argument('group_b', type=int, help='second Telegram group id')
    return parser.parse_args(argv)


def main(argv=None):
    args = parse_args(argv)

    # Lazy imports: keep `--help` working when REDIS_URL is unset. Redis still
    # supplies the Telegram config and login session; no export is written.
    from common import member_dict
    from config import load_app_config
    from crawl import fetch_group, login

    client = login(load_app_config())

    groups = []
    for group_id in (args.group_a, args.group_b):
        try:
            entity, members = fetch_group(client, group_id, 'fetching')
        except Exception as e:
            print(f'error accessing group {group_id}: {e}', file=sys.stderr)
            return 1
        if entity is None:
            print(f'{group_id} is not a group.', file=sys.stderr)
            return 1
        groups.append({
            'group_id': group_id,
            'title': entity.title,
            'members': [member_dict(m) for m in members],
        })

    group_a, group_b = groups
    only_a, only_b, both = split_members(group_a, group_b)
    print()
    print_summary(group_a, group_b, only_a, only_b, both)
    return 0


def split_members(group_a, group_b):
    """Return (only in A, only in B, in both) by Telegram member id."""
    a_members = {m['id']: m for m in group_a['members']}
    b_members = {m['id']: m for m in group_b['members']}
    only_a = [a_members[i] for i in sorted(a_members.keys() - b_members.keys())]
    only_b = [b_members[i] for i in sorted(b_members.keys() - a_members.keys())]
    both = [b_members[i] for i in sorted(a_members.keys() & b_members.keys())]
    return only_a, only_b, both


def group_label(group):
    return f'{group["title"]} ({group["group_id"]})'


def print_summary(group_a, group_b, only_a, only_b, both):
    print(f'diff --telegram-groups "{group_label(group_a)}" "{group_label(group_b)}"')
    print(color(f'--- {group_label(group_a)}', RED))
    print(color(f'+++ {group_label(group_b)}', GREEN))
    print(color(
        f'@@ members: {len(group_a["members"])} vs {len(group_b["members"])} '
        f'(both {len(both)}, -{len(only_a)} only first, +{len(only_b)} only second) @@',
        CYAN,
    ))
    print()

    widths = column_widths([*only_a, *only_b, *both])
    for member in both:
        print(format_member_row(' ', member, widths))
    for member in only_a:
        print(color(format_member_row('-', member, widths), RED))
    for member in only_b:
        print(color(format_member_row('+', member, widths), GREEN))

    if not group_a['members'] and not group_b['members']:
        print(' both groups have no visible members.')


if __name__ == '__main__':
    raise SystemExit(main())
