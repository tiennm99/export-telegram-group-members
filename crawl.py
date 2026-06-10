import getpass

from telethon.errors import SessionPasswordNeededError
from telethon.sessions import StringSession
from telethon.sync import TelegramClient
from telethon.tl.types import Chat, Channel

from common import new_run_time, save_group_export
from config import load_app_config
from session_store import load_session, save_session


def main():
    app_config = load_app_config()
    api_hash = app_config['api_hash']
    api_id = app_config['api_id']
    group_ids = app_config['group_ids']
    phone = app_config['phone']

    # Session loads from Redis: a saved string means no re-login on any device.
    client = TelegramClient(StringSession(load_session()), api_id, api_hash)

    client.connect()
    if not client.is_user_authorized():
        client.send_code_request(phone)
        try:
            client.sign_in(code=input('Enter code: '))
        except SessionPasswordNeededError:
            client.sign_in(password=getpass.getpass())

    save_session(client.session.save())

    run_time = new_run_time()
    saved = 0
    for group_id in group_ids:
        try:
            entity = client.get_entity(group_id)
            if not isinstance(entity, (Chat, Channel)):
                print(f'skip {group_id} because it is not a group.')
                continue
            print(f'exporting {entity.title} (ID: {group_id})')
            members = client.get_participants(entity)
            members.sort(key=lambda x: x.id)
            save_group_export(group_id, entity.title, members, run_time)
            saved += 1
            print(f'export {entity.title} done.')
        except Exception as e:
            print(f'error accessing group {group_id}: {e}')

    print(f'saved run {run_time}: {saved} group(s) to Redis.')
    return 0


if __name__ == '__main__':
    raise SystemExit(main())
