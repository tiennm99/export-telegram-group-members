import os
import json

import redis
from dotenv import load_dotenv

load_dotenv()

# Shared Redis holds the Telegram session + export history so the tool runs on
# any device from just .env. The prefix is fixed to keep local setup minimal.
redis_url = os.getenv('REDIS_URL')
if not redis_url:
    raise SystemExit('REDIS_URL not set in .env (e.g. rediss://default:<password>@<host>:<port>)')

redis_prefix = 'telegram-export'
redis_client = redis.from_url(redis_url, decode_responses=True)


def key(*parts):
    """Build a namespaced Redis key. Single source of the prefix (DRY)."""
    return ':'.join([redis_prefix, *parts])


def load_app_config():
    """Load Telegram crawl config from Redis."""
    raw = redis_client.get(key('config'))
    if not raw:
        raise SystemExit(
            'telegram config not found in Redis. run configure.py first.'
        )
    try:
        config = json.loads(raw)
    except ValueError as exc:
        raise SystemExit('telegram config in Redis is not valid JSON.') from exc

    missing = [
        name for name in ('api_id', 'api_hash', 'phone', 'group_ids')
        if not config.get(name)
    ]
    if missing:
        raise SystemExit(f'telegram config missing: {", ".join(missing)}')

    return {
        'api_id': int(config['api_id']),
        'api_hash': str(config['api_hash']),
        'phone': str(config['phone']),
        'group_ids': [int(group_id) for group_id in config['group_ids']],
    }


def save_app_config(api_id, api_hash, phone, group_ids):
    """Save Telegram crawl config to Redis."""
    record = {
        'api_id': int(api_id),
        'api_hash': str(api_hash),
        'phone': str(phone),
        'group_ids': [int(group_id) for group_id in group_ids],
    }
    redis_client.set(key('config'), json.dumps(record, ensure_ascii=False))
    return record
