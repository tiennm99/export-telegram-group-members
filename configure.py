import argparse
import getpass


def parse_args():
    parser = argparse.ArgumentParser(
        description='Store Telegram crawl config in Redis.',
    )
    parser.add_argument('--api-id', required=True, type=int, help='Telegram API ID')
    parser.add_argument('--api-hash', help='Telegram API hash')
    parser.add_argument('--phone', required=True, help='Telegram account phone')
    parser.add_argument(
        '--groups',
        required=True,
        help='Comma-separated Telegram group IDs to crawl',
    )
    return parser.parse_args()


def parse_group_ids(value):
    group_ids = [part.strip() for part in value.split(',') if part.strip()]
    if not group_ids:
        raise SystemExit('at least one group id is required')
    return [int(group_id) for group_id in group_ids]


def main():
    args = parse_args()
    group_ids = parse_group_ids(args.groups)
    api_hash = args.api_hash or getpass.getpass('API hash: ')
    if not api_hash:
        raise SystemExit('api hash is required')

    from config import save_app_config

    config = save_app_config(args.api_id, api_hash, args.phone, group_ids)
    print(f'saved Redis config for {config["phone"]}: {len(group_ids)} group(s).')
    return 0


if __name__ == '__main__':
    raise SystemExit(main())
