"""Migrate old phone-scoped Redis keys to the current phone-less key format.

Old:
    telegram-export:session:<phone>
    telegram-export:run:<phone>:<yyyymmddhhmmss>:<group_id>

New:
    telegram-export:session
    telegram-export:run:<yyyymmddhhmmss>:<group_id>
"""

import argparse
import json
import os

import redis
from dotenv import load_dotenv

PREFIX = 'telegram-export'


def redis_key(*parts):
    return ':'.join([PREFIX, *parts])


def parse_args():
    parser = argparse.ArgumentParser(
        description='Migrate Redis keys from phone-scoped format to phone-less format.',
    )
    parser.add_argument(
        '--phone',
        help='Phone part in old Redis keys. Defaults to telegram-export:config phone.',
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
        help='Delete old phone-scoped keys after successful copy.',
    )
    return parser.parse_args()


def connect_redis():
    load_dotenv()
    redis_url = os.getenv('REDIS_URL')
    if not redis_url:
        raise SystemExit('REDIS_URL not set in .env')
    return redis.from_url(redis_url, decode_responses=True)


def resolve_phone(client, phone):
    if phone:
        return phone

    raw_config = client.get(redis_key('config'))
    if not raw_config:
        raise SystemExit('No config found. Pass --phone explicitly.')

    try:
        config = json.loads(raw_config)
    except ValueError as exc:
        raise SystemExit('telegram-export:config is not valid JSON.') from exc

    config_phone = config.get('phone')
    if not config_phone:
        raise SystemExit('Config has no phone. Pass --phone explicitly.')
    return str(config_phone)


def migrate_key(client, old_key, new_key, args):
    value = client.get(old_key)
    if value is None:
        return 'missing'

    if client.exists(new_key) and not args.overwrite:
        print(f'skip existing: {new_key}')
        return 'skipped'

    print(f'{old_key} -> {new_key}')
    if args.dry_run:
        return 'planned'

    client.set(new_key, value)
    if args.delete_old:
        client.delete(old_key)
    return 'migrated'


def migrate_session(client, phone, args):
    return migrate_key(
        client,
        redis_key('session', phone),
        redis_key('session'),
        args,
    )


def migrate_runs(client, phone, args):
    counts = {'planned': 0, 'migrated': 0, 'skipped': 0, 'missing': 0}
    pattern = redis_key('run', phone, '*')

    for old_key in client.scan_iter(match=pattern):
        parts = old_key.split(':')
        if len(parts) != 5:
            print(f'skip invalid old run key: {old_key}')
            counts['skipped'] += 1
            continue

        _, _, _, run_time, group_id = parts
        result = migrate_key(
            client,
            old_key,
            redis_key('run', run_time, group_id),
            args,
        )
        counts[result] += 1

    return counts


def main():
    args = parse_args()
    client = connect_redis()
    phone = resolve_phone(client, args.phone)

    print(f'migrating phone-scoped keys for {phone}')
    session_result = migrate_session(client, phone, args)
    run_counts = migrate_runs(client, phone, args)

    print('')
    print(f'session: {session_result}')
    print(
        'runs: '
        f'{run_counts["migrated"]} migrated, '
        f'{run_counts["planned"]} planned, '
        f'{run_counts["skipped"]} skipped'
    )
    if args.dry_run:
        print('dry run only; no Redis keys changed.')


if __name__ == '__main__':
    raise SystemExit(main())
