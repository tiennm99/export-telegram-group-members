import os

from dotenv import load_dotenv

load_dotenv()

phone = os.getenv('PHONE')
api_id = os.getenv('API_ID')
api_hash = os.getenv('API_HASH')

proxy_host = os.getenv('PROXY_HOST')
proxy_port = int(os.getenv('PROXY_PORT', '0')) if os.getenv('PROXY_PORT') else None
proxy_secret = os.getenv('PROXY_SECRET')

group_names_to_export = ['ZingPlay Game Studios', 'ZPS HCM', 'ZPS HCM - Xin nghỉ (phép/đi trễ)']  # Add the group names you want to export here
