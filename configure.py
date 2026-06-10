import argparse
import getpass


def parse_args():
    parser = argparse.ArgumentParser(
        description='Store Telegram crawl config in Redis.',
    )
    parser.add_argument('--api-id', type=int, help='Telegram API ID')
    parser.add_argument('--api-hash', help='Telegram API hash')
    parser.add_argument('--phone', help='Telegram account phone')
    parser.add_argument(
        '--groups',
        help='Comma-separated Telegram group IDs to crawl',
    )
    return parser.parse_args()


def prompt_required(label):
    value = input(f'{label}: ').strip()
    if not value:
        raise SystemExit(f'{label.lower()} is required')
    return value


def prompt_api_id(value):
    if value is not None:
        return value
    try:
        return int(prompt_required('API ID'))
    except ValueError as exc:
        raise SystemExit('api id must be an integer') from exc


def prompt_api_hash(value):
    api_hash = value or getpass.getpass('API hash: ')
    if not api_hash:
        raise SystemExit('api hash is required')
    return api_hash


def prompt_phone(value):
    return value or prompt_required('Phone')


def prompt_groups(value):
    return parse_group_ids(value or prompt_required('Group IDs'))


def parse_group_ids(value):
    group_ids = [part.strip() for part in value.split(',') if part.strip()]
    if not group_ids:
        raise SystemExit('at least one group id is required')
    try:
        return [int(group_id) for group_id in group_ids]
    except ValueError as exc:
        raise SystemExit('group ids must be comma-separated integers') from exc


def main():
    args = parse_args()
    api_id = prompt_api_id(args.api_id)
    api_hash = prompt_api_hash(args.api_hash)
    phone = prompt_phone(args.phone)
    group_ids = prompt_groups(args.groups)

    # Lazy import: keeps `python configure.py --help` working when REDIS_URL is unset.
    from config import save_app_config

    config = save_app_config(api_id, api_hash, phone, group_ids)
    print(f'saved Redis config for {config["phone"]}: {len(group_ids)} group(s).')
    return 0


if __name__ == '__main__':
    raise SystemExit(main())
