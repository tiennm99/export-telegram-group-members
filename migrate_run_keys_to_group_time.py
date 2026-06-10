"""Migrate Redis run keys from run:<time>:<group_id> to run:<group_id>:<time>."""

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
        description='Migrate Redis export run keys to group-first format.',
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
        help='Delete old run:<time>:<group_id> keys after successful copy.',
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


def migrate_run_key(client, old_key, args):
    parts = old_key.split(':')
    if len(parts) != 4:
        return 'skipped'

    _, _, run_time, group_id = parts
    if not is_run_time(run_time):
        return 'skipped'

    new_key = redis_key('run', group_id, run_time)
    if client.exists(new_key) and not args.overwrite:
        if args.delete_old and not args.dry_run:
            client.delete(old_key)
            print(f'deleted old existing: {old_key}')
            return 'deleted'
        print(f'skip existing: {new_key}')
        return 'skipped'

    value = client.get(old_key)
    if value is None:
        return 'missing'

    print(f'{old_key} -> {new_key}')
    if args.dry_run:
        return 'planned'

    client.set(new_key, value)
    if args.delete_old:
        client.delete(old_key)
    return 'migrated'


def main():
    args = parse_args()
    client = connect_redis()
    counts = {'planned': 0, 'migrated': 0, 'deleted': 0, 'skipped': 0, 'missing': 0}

    for old_key in client.scan_iter(match=redis_key('run', '*', '*')):
        result = migrate_run_key(client, old_key, args)
        counts[result] += 1

    print('')
    print(
        'runs: '
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
