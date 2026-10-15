import argparse
import getpass

from telethon.errors import SessionPasswordNeededError
from telethon.sessions import StringSession
from telethon.sync import TelegramClient
from telethon.tl.types import Chat, Channel

from common import new_run_time, save_group_export
from config import load_app_config
from session_store import load_session, save_session


def parse_args(argv=None):
    parser = argparse.ArgumentParser(
        description='Export members from Telegram groups.',
    )
    parser.add_argument(
        'group_ids',
        nargs='*',
        type=int,
        metavar='group_id',
        help='Telegram group ID to crawl; defaults to the configured group IDs',
    )
    return parser.parse_args(argv)


def login(app_config):
    """Connect to Telegram, signing in only when the stored session is missing."""
    # Session loads from Redis: a saved string means no re-login on any device.
    client = TelegramClient(
        StringSession(load_session()),
        app_config['api_id'],
        app_config['api_hash'],
    )

    client.connect()
    if not client.is_user_authorized():
        client.send_code_request(app_config['phone'])
        try:
            client.sign_in(code=input('Enter code: '))
        except SessionPasswordNeededError:
            client.sign_in(password=getpass.getpass())

    save_session(client.session.save())
    return client


def fetch_group(client, group_id, action):
    """Return (entity, members sorted by id), or (None, []) for a non-group."""
    entity = client.get_entity(group_id)
    if not isinstance(entity, (Chat, Channel)):
        return None, []
    print(f'{action} {entity.title} (ID: {group_id})')
    members = client.get_participants(entity)
    members.sort(key=lambda x: x.id)
    return entity, members


def main(argv=None):
    args = parse_args(argv)
    app_config = load_app_config()
    # Command-line IDs override the configured list for one-off crawls.
    group_ids = args.group_ids or app_config['group_ids']

    client = login(app_config)

    run_time = new_run_time()
    saved = 0
    for group_id in group_ids:
        try:
            entity, members = fetch_group(client, group_id, 'exporting')
            if entity is None:
                print(f'skip {group_id} because it is not a group.')
                continue
            save_group_export(group_id, entity.title, members, run_time)
            saved += 1
            print(f'export {entity.title} done.')
        except Exception as e:
            print(f'error accessing group {group_id}: {e}')

    print(f'saved run {run_time}: {saved} group(s) to Redis.')
    return 0


if __name__ == '__main__':
    raise SystemExit(main())
