"""Migrate Redis export keys from run:*:* to group:<group_id>:<time>.

Supported source formats:
    run:<time>:<group_id>
    run:<group_id>:<time>
"""

import argparse
import os

import redis
from dotenv import load_dotenv

PREFIX = 'telegram-export'
TIME_LENGTH = 14


def redis_key(*parts):
    return ':'.join([PREFIX, *parts])


def parse_args():
    parser = argparse.ArgumentParser(
        description='Migrate Redis export run keys to group key format.',
    )
    parser.add_argument(
        '--dry-run',
        action='store_true',
        help='Print planned changes without writing Redis.',
    )
    parser.add_argument(
        '--overwrite',
        action='store_true',
        help='Overwrite new-format keys when they already exist.',
    )
    parser.add_argument(
        '--delete-old',
        action='store_true',
        help='Delete old run:*:* keys after successful copy.',
    )
    return parser.parse_args()


def connect_redis():
    load_dotenv()
    redis_url = os.getenv('REDIS_URL')
    if not redis_url:
        raise SystemExit('REDIS_URL not set in .env')
    return redis.from_url(redis_url, decode_responses=True)


def is_run_time(value):
    return len(value) == TIME_LENGTH and value.isdigit()


def export_key_parts(export_key):
    parts = export_key.split(':')
    if len(parts) != 4:
        return None

    _, key_type, first, second = parts
    if key_type != 'run':
        return None
    if is_run_time(first) and not is_run_time(second):
        return second, first
    if not is_run_time(first) and is_run_time(second):
        return first, second
    return None


def migrate_export_key(client, old_key, args):
    key_parts = export_key_parts(old_key)
    if key_parts is None:
        return 'skipped'

    group_id, run_time = key_parts
    new_key = redis_key('group', group_id, run_time)

    value = client.get(old_key)
    if value is None:
        return 'missing'

    if client.exists(new_key) and not args.overwrite:
        if args.delete_old and not args.dry_run:
            client.delete(old_key)
            print(f'deleted old existing: {old_key}')
            return 'deleted'
        print(f'skip existing: {new_key}')
        return 'skipped'

    print(f'{old_key} -> {new_key}')
    if args.dry_run:
        return 'planned'

    client.set(new_key, value)
    if args.delete_old:
        client.delete(old_key)
    return 'migrated'


def migrate_run_key(client, old_key, args):
    """Compatibility wrapper for older imports/tests."""
    return migrate_export_key(client, old_key, args)


def main():
    args = parse_args()
    client = connect_redis()
    counts = {'planned': 0, 'migrated': 0, 'deleted': 0, 'skipped': 0, 'missing': 0}

    for old_key in client.scan_iter(match=redis_key('run', '*', '*')):
        result = migrate_export_key(client, old_key, args)
        counts[result] += 1

    print('')
    print(
        'exports: '
        f'{counts["migrated"]} migrated, '
        f'{counts["deleted"]} deleted, '
        f'{counts["planned"]} planned, '
        f'{counts["skipped"]} skipped, '
        f'{counts["missing"]} missing'
    )
    if args.dry_run:
        print('dry run only; no Redis keys changed.')


if __name__ == '__main__':
    raise SystemExit(main())
