import getpass

from telethon.errors import SessionPasswordNeededError
from telethon.sync import TelegramClient
from telethon.tl.types import Chat, Channel

from common import *
from config import *

client = TelegramClient(phone, api_id, api_hash)

client.connect()
if not client.is_user_authorized():
    client.send_code_request(phone)
    try:
        client.sign_in(code=input('Enter code: '))
    except SessionPasswordNeededError:
        client.sign_in(password=getpass.getpass())

client.start(phone)

for group_id in group_ids:
    try:
        entity = client.get_entity(group_id)
        if not isinstance(entity, (Chat, Channel)):
            print(f'skip {group_id} because it is not a group.')
            continue
        print(f'exporting {entity.title} (ID: {group_id})')
        members = client.get_participants(entity)
        members.sort(key=lambda x: x.id)
        export_csv(entity.title, members)
        print(f'export {entity.title} done.')
    except Exception as e:
        print(f'error accessing group {group_id}: {e}')
