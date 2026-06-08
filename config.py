import os

import redis
from dotenv import load_dotenv

load_dotenv()

phone = os.getenv('PHONE')
api_id = os.getenv('API_ID')
api_hash = os.getenv('API_HASH')
group_ids_str = os.getenv('GROUP_IDS', '')
group_ids = [int(id.strip()) for id in group_ids_str.split(',') if id.strip()] if group_ids_str else []

# Shared Redis holds the Telegram session + export history so the tool runs on
# any device from just .env. REDIS_PREFIX namespaces every key so this project
# never collides with others sharing the same Redis instance.
redis_url = os.getenv('REDIS_URL')
if not redis_url:
    raise SystemExit('REDIS_URL not set in .env (e.g. rediss://default:<password>@<host>:<port>)')

redis_prefix = os.getenv('REDIS_PREFIX', 'telegram-export')
redis_client = redis.from_url(redis_url, decode_responses=True)


def key(*parts):
    """Build a namespaced Redis key. Single source of the prefix (DRY)."""
    return ':'.join([redis_prefix, *parts])
