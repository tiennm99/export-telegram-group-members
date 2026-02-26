import os

from dotenv import load_dotenv

load_dotenv()

phone = os.getenv('PHONE')
api_id = os.getenv('API_ID')
api_hash = os.getenv('API_HASH')
group_ids_str = os.getenv('GROUP_IDS', '')
group_ids = [int(id.strip()) for id in group_ids_str.split(',') if id.strip()] if group_ids_str else []
