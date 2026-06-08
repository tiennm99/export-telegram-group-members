"""Load/save the Telethon StringSession from Redis.

The session string grants full account access; it lives only in Redis (behind
TLS/auth), never on local disk. Deleting the key simply forces a clean re-login.
"""

from config import key, phone, redis_client

_SESSION_KEY = key('session', phone)


def load_session():
    """Return the stored StringSession string, or None if not yet authenticated."""
    return redis_client.get(_SESSION_KEY)


def save_session(session_string):
    """Persist the StringSession string for reuse on any device."""
    redis_client.set(_SESSION_KEY, session_string)
