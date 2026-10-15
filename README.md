# export-telegram-group-members

Export Telegram group members with Go and [gotd/td](https://github.com/gotd/td).
Admin rights give full member visibility. The login session and export history
live in **Redis**, so you can run the tool on any device from just a `.env`
file: no re-login and no copying session or CSV files.

## How to use

1. Clone this repository and install [Go](https://go.dev/dl/):

```bash
git clone https://github.com/tiennm99/export-telegram-group-members.git
```

2. Create a new Telegram app at [https://my.telegram.org](https://my.telegram.org) and get the `api_id` and `api_hash`.
3. Create a free Redis database (e.g. [Upstash](https://upstash.com)) and copy its `rediss://` connection URL.
4. Copy `.env.example` to `.env` and fill in `REDIS_URL`.
5. Store Telegram config in Redis:

```bash
go run ./cmd/configure
```

`configure` prompts for `api_id`, `api_hash`, phone, and group IDs. You can
pass any of them as flags instead: `--api-id`, `--api-hash`, `--phone`, and
`--groups -1001234567890,-1009876543210`. All commands read the stored group
IDs from this Redis config.

6. Crawl the configured groups:

```bash
go run ./cmd/crawl
```

Every group ID stored in the Redis config is crawled. To change the list, run
`configure` again.

To crawl other groups just once, pass their IDs; they replace the configured
list for that run only:

```bash
go run ./cmd/crawl -1001234567890 -1009876543210
```

The first crawl asks for the Telegram login code once, then stores the session
in Redis. Any later run on any device pointed at the same Redis reuses the
Redis config and session and **does not** prompt again.

Group IDs use the usual marked form: `-100<id>` for supergroups and channels,
`-<id>` for basic groups. The account must already be a member of each group.

## Compare two crawls

Compare membership changes for one group between two saved crawls with
git-diff-style terminal output:

```bash
go run ./cmd/compare <group_id> <time1> <time2>
```

If `time1` and `time2` are omitted, the command compares the latest two crawls
for that group:

```bash
go run ./cmd/compare <group_id>
```

If `group_id` is also omitted, the command uses the first group in the stored
configuration:

```bash
go run ./cmd/compare
```

## Compare two groups

Fetch two groups live from Telegram and compare their members by Telegram ID:

```bash
go run ./cmd/compare-groups <group_id_a> <group_id_b>
```

Members in both groups are listed unprefixed, members only in the first group
with `-`, and members only in the second group with `+`. No crawl is saved to
Redis; the command uses Redis only for the Telegram config and login session,
the same way `crawl` does.

## Export a crawl to CSV

Export one group's saved crawl from Redis:

```bash
go run ./cmd/export [group_id] [timecrawl]
```

When `timecrawl` is omitted, the latest crawl for the group is exported. When
both arguments are omitted, the first group in the stored configuration and its
latest crawl are used. An explicit crawl time must use `yyyymmddhhmmss` format:

```bash
go run ./cmd/export -1001234567890 20260724120000
```

CSV files are written to:

```text
output/<group-id>-<yyyymmddhhmmss>.csv
```

Each file contains the columns `id`, `username`, `first_name`, and `last_name`.
Formula-like Telegram text is prefixed with an apostrophe for safe spreadsheet
opening. On POSIX systems, the `output/` directory is created with `0700`
permissions and each CSV file is written with `0600` permissions. The `output/`
directory is ignored by Git.

## Build binaries

The code is pure Go, so static binaries build without cgo:

```bash
CGO_ENABLED=0 go build -o bin/ ./cmd/...
```

## Configuration

| Variable | Description |
|----------|-------------|
| `REDIS_URL` | Redis connection string (`rediss://default:<password>@<host>:<port>`) |

Telegram `api_id`, `api_hash`, `phone`, and `group_ids` are stored in Redis by `configure`.

## How data is stored

All data lives in Redis under the `telegram-export` prefix. No key references another, so deleting any key can never corrupt another's state:

```
telegram-export:config                         -> JSON config:
                                 { api_id, api_hash, phone, group_ids }
telegram-export:session-gotd                   -> gotd session JSON (login)
telegram-export:group:<group_id>:<yyyymmddhhmmss> -> one group's export as JSON:
                                 { group_id, title, time,
                                   members: [{ id, username, first_name, last_name }] }
```

All groups in one crawl share one timestamp. The login session grants full
account access, so restrict Redis access accordingly; deleting
`telegram-export:session-gotd` forces a clean re-login.

## Development

```bash
go vet ./...
go test ./...
```

Tests run offline: Redis is faked with miniredis, and Telegram calls are not
exercised.
