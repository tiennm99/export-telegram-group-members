import getpass

from telethon.errors import SessionPasswordNeededError
from telethon.sync import TelegramClient
import TelethonFakeTLS

from common import *
from config import *

if proxy_host and proxy_port and proxy_secret:
    proxy = (proxy_host, proxy_port, proxy_secret)
    connection = TelethonFakeTLS.ConnectionTcpMTProxyFakeTLS
    client = TelegramClient(phone, api_id, api_hash, connection=connection, proxy=proxy)
else:
    client = TelegramClient(phone, api_id, api_hash)


client.connect()
if not client.is_user_authorized():
    client.send_code_request(phone)
    try:
        client.sign_in(code=input('Enter code: '))
    except SessionPasswordNeededError:
        client.sign_in(password=getpass.getpass())

client.start(phone)

dialogs = client.get_dialogs()
for dialog in dialogs:
    if not dialog.is_group:
        print('skip ' + dialog.name + ' because it is not a group.')
        continue
    if dialog.name not in group_names_to_export:
        print('skip ' + dialog.name + ' because it is not in the list of group names to export.')
        continue
    print('exporting ' + dialog.name)
    members = client.get_participants(dialog)
    members.sort(key=lambda x: x.id)
    export_csv(dialog.name, members)
    print('export ' + dialog.name + ' done.')
