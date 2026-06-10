# export-telegram-group-members

Export Telegram group members using Telethon — admin auth required for full member visibility. Session and export history are stored in **Redis**, so you can run the tool on any device from just a `.env` file (no re-login, no copying session/CSV files).

## How to use

1. Clone this repository:

```bash
git clone https://github.com/tiennm99/export-telegram-group-members.git
```

2. Install requirements:

```bash
pip install -r requirements.txt
```

3. Create a new Telegram app at [https://my.telegram.org](https://my.telegram.org) and get the `api_id` and `api_hash`.
4. Create a free Redis database (e.g. [Upstash](https://upstash.com)) and copy its `rediss://` connection URL.
5. Copy `.env.example` to `.env` and fill in `REDIS_URL`.
6. Store Telegram config in Redis:

```bash
python configure.py
```

`configure.py` prompts for `api_id`, `api_hash`, phone, and group IDs.

7. Crawl configured groups:

```bash
python crawl.py
```

The first crawl asks for the Telegram login code once, then stores the session in Redis. Any later run — on any device pointed at the same Redis — reuses the Redis config and session, and **does not** prompt again.

## Compare two crawls

Compare membership changes for one group between two saved crawls with the
git-diff-style terminal output:

```bash
python compare.py <group_id> <time1> <time2>
```

If `time1` and `time2` are omitted, the command compares the latest two crawls
for that group:

```bash
python compare.py <group_id>
```

## Configuration

| Variable | Description |
|----------|-------------|
| `REDIS_URL` | Redis connection string (`rediss://default:<password>@<host>:<port>`) |

Telegram `api_id`, `api_hash`, `phone`, and `group_ids` are stored in Redis by `configure.py`.

## How data is stored

All data lives in Redis under the `telegram-export` prefix. No key references another, so deleting any key can never corrupt another's state:

```
telegram-export:config                         -> JSON config:
                                 { api_id, api_hash, phone, group_ids }
telegram-export:session                        -> StringSession string (login)
telegram-export:group:<group_id>:<yyyymmddhhmmss> -> one group's export as JSON:
                                 { group_id, title, time,
                                   members: [{ id, username, first_name, last_name }] }
```

Each group is written as its own key per run; all groups in one run share the same `yyyymmddhhmmss` timestamp. Read the history programmatically:

```python
from common import list_exports

for rec in list_exports():       # sorted by (time, group_id)
    print(rec['time'], rec['group_id'], rec['title'], len(rec['members']), 'members')
```

## Rate limits and visibility notes

- Telegram limits how fast you can fetch participants; large groups may take longer.
- For **supergroups**, only admins can retrieve the full member list — regular members see a partial list or get an error.
- For private groups where you are not a member, access will be denied.

## Security

- The session string grants **full access to your Telegram account**. It lives only in Redis (use a TLS `rediss://` URL) and is never written to disk or committed to git.
- `.env` is git-ignored. Never commit your `REDIS_URL` or session string.

## License

Apache-2.0 — see [LICENSE](LICENSE).
