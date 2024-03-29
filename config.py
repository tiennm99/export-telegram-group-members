import os

from dotenv import load_dotenv

load_dotenv()

phone = os.getenv('PHONE')
api_id = os.getenv('API_ID')
api_hash = os.getenv('API_HASH')

group_names_to_export = ['sample1', 'sample2']  # Add the group names you want to export here
