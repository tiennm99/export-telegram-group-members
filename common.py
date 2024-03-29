import csv
import os
import re
from datetime import datetime

now = datetime.now()
folder_name = now.strftime('%Y-%m-%d %H-%M-%S')
os.mkdir(folder_name)


def remove_special_characters(string):
    return re.sub('[<>:"/\\\\|?*]', '', string)


def export_csv(group_name, members):
    filename = remove_special_characters(group_name) + '.csv'
    path = os.path.join(folder_name, filename)
    with open(path, 'w', encoding='UTF-8') as f:
        writer = csv.writer(f, delimiter=',', lineterminator='\n')
        writer.writerow(['id', 'username', 'first_name', 'last_name'])
        for member in members:
            writer.writerow([member.id, member.username, member.first_name, member.last_name])
