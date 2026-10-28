# PG-BKUP

**PG-BKUP** backs up, restores and migrates **PostgreSQL** databases. It streams `pg_dump` output through compression and encryption straight into the storage: no temporary dump file, constant memory, whatever the database size.

[![Tests](https://github.com/jkaninda/pg-bkup/actions/workflows/tests.yml/badge.svg)](https://github.com/jkaninda/pg-bkup/actions/workflows/tests.yml)
[![Build](https://github.com/jkaninda/pg-bkup/actions/workflows/release.yml/badge.svg)](https://github.com/jkaninda/pg-bkup/actions/workflows/release.yml)
[![Go Report](https://goreportcard.com/badge/github.com/jkaninda/pg-bkup)](https://goreportcard.com/report/github.com/jkaninda/pg-bkup)
![Docker Image Size (latest by date)](https://img.shields.io/docker/image-size/jkaninda/pg-bkup?style=flat-square)
![Docker Pulls](https://img.shields.io/docker/pulls/jkaninda/pg-bkup?style=flat-square)
<a href="https://ko-fi.com/jkaninda"><img src="https://uploads-ssl.webflow.com/5c14e387dab576fe667689cf/5cbed8a4ae2b88347c06c923_BuyMeACoffee_blue.png" height="20" alt="buy ma a coffee"></a>

Documentation: <https://jkaninda.github.io/pg-bkup> · MySQL: [mysql-bkup](https://github.com/jkaninda/mysql-bkup)

## Features

- **Storage:** local, Amazon S3 and S3-compatible (MinIO, Wasabi, R2...), Google Cloud Storage, Azure Blob, SFTP, FTP/FTPS.
- **Streaming:** `pg_dump | gzip/zstd | GPG/age | upload`, and back for restores. Nothing is written to disk.
- **Encryption:** GPG (passphrase or public key) and [age](https://age-encryption.org) (recipients or passphrase).
- **Safe restores:** `psql -v ON_ERROR_STOP=1 --single-transaction` (or `pg_restore --single-transaction`): a failing or corrupted backup changes nothing.
- **Every PostgreSQL version:** the image ships the 16, 17 and 18 clients and picks, per server, the lowest `pg_dump` that is not older than the server (servers 9.2 and newer). No more "server version mismatch".
- **Backup index and retention:** by age (`BACKUP_RETENTION_DAYS`) or count (`BACKUP_RETENTION_COUNT`); only backups recorded in the index are ever deleted.
- **Commands:** `backup`, `restore`, `migrate`, `verify`, `list`, `delete`, `prune`, `reconcile`, `config`, `health`.
- **Scheduling:** `--cron-expression` daemon with a Docker `HEALTHCHECK` and optional Prometheus metrics.
- **Notifications:** email, Telegram, Slack, Discord, ntfy, generic webhook, healthchecks.io pings.
- **Formats:** plain SQL (default, `.sql`) or custom (`PG_DUMP_FORMAT=custom`, `.dump`); schema-only, data-only, selected or excluded tables, all databases one file each, or the whole instance with `pg_dumpall`.

## Quick start

```shell
docker run --rm --network your_network \
  -v $PWD/backup:/backup \
  -e DB_HOST=postgres -e DB_USER=postgres -e DB_PASSWORD=password \
  jkaninda/pg-bkup backup -d database_name
```

The last line of a successful backup always names the artifact:

```
level=INFO msg="Backup completed" database=database_name file=database_name_20261027_031500.sql.gz size=12.3MiB duration=2.1s storage=local location=/backup
```

Restore it (the target database must exist):

```shell
docker run --rm --network your_network \
  -v $PWD/backup:/backup \
  -e DB_HOST=postgres -e DB_USER=postgres -e DB_PASSWORD=password \
  jkaninda/pg-bkup restore -d database_name -f database_name_20261027_031500.sql.gz
```

`restore` accepts `.sql`, `.sql.gz`, `.sql.zst`, `.dump` (custom format) and any of them encrypted (`.gpg`, `.age`). The format is detected from the content, not the name.

A connection URL works too: `DB_URL=postgres://user:p%40ss@host:5432/db?sslmode=require` or the 1.x JDBC form `jdbc:postgresql://host:5432/db?user=user&password=pass`.

### Backup variants

| Command | File name |
|---|---|
| `backup -d shop` | `shop_20261027_031500.sql.gz` |
| `backup -d shop --schema-only` | `shop_schema_20261027_031500.sql.gz` |
| `backup -d shop --data-only` | `shop_data_20261027_031500.sql.gz` |
| `backup -d shop -t users -t orders` (or `--tables users,orders`) | `shop_tables_2_20261027_031500.sql.gz` |
| `backup -d shop --exclude-table audit_log` | `shop_20261027_031500.sql.gz` |
| `backup -d shop --custom-name nightly` | `nightly.sql.gz` (overwritten each run) |
| `PG_DUMP_FORMAT=custom backup -d shop` | `shop_20261027_031500.dump.gz` |
| `backup --all-databases --exclude-db _aiven,defaultdb` | one file per database |
| `backup --all-in-one --no-role-passwords` | `all_databases_20261027_031500.sql.gz` (`pg_dumpall`: roles, tablespaces, every database) |

`--all-databases` skips templates and the `postgres` database; databases are listed from `DB_AUTH_DATABASE` (default `postgres`).
An `--all-in-one` file is restored with `restore -f all_databases_....sql.gz` (no `-d`): it is replayed without a single transaction and, by default, without `ON_ERROR_STOP`, because `pg_dumpall` output always contains statements such as `CREATE ROLE postgres` that fail on an existing server (`RESTORE_ON_ERROR_STOP`).

The custom format is dumped with `pg_dump -Fc -Z0` and compressed by the pipeline (parallel gzip or zstd), so `COMPRESSION` applies to both formats and every artifact is decompressed the same way. Custom-format archives are restored with `pg_restore --single-transaction --exit-on-error --no-owner` (`RESTORE_NO_OWNER=false` keeps ownership).

### Migrate

`migrate` pipes `pg_dump` into `psql` on the target (`TARGET_DB_*`), with no intermediate file:

```shell
docker run --rm --network your_network \
  -e DB_HOST=old -e DB_USER=postgres -e DB_PASSWORD=pw -e DB_NAME=shop \
  -e TARGET_DB_HOST=new -e TARGET_DB_USER=postgres -e TARGET_DB_PASSWORD=pw -e TARGET_DB_NAME=shop \
  jkaninda/pg-bkup migrate
```

- `migrate --all-databases` (`-a`) copies every database into the same name, creating missing ones.
- `migrate --entire-instance` (`-I`) copies roles, tablespaces and every database with `pg_dumpall` (add `--no-role-passwords` to keep the target's passwords).
- A single database is copied in one transaction; the target database is created if it does not exist.
- `TARGET_DB_URL`, `TARGET_DB_SSL_MODE`, `TARGET_DB_SSL_CA/CERT/KEY` and `TARGET_DB_AUTH_DATABASE` are honoured.

## Storage

```shell
# S3 / MinIO
-e AWS_S3_ENDPOINT=https://s3.amazonaws.com -e AWS_S3_BUCKET_NAME=backups -e AWS_REGION=us-east-1 \
-e AWS_ACCESS_KEY=... -e AWS_SECRET_KEY=... \
jkaninda/pg-bkup backup --storage s3 -d shop --path /prod/shop
# MinIO: add AWS_DISABLE_SSL=true (plain http) and AWS_FORCE_PATH_STYLE=true.
# Without keys, the AWS default credential chain (IAM role, IRSA, profile) is used.

# SFTP (host keys are verified: mount a known_hosts file)
-e SFTP_HOST=backup.example.com -e SFTP_USER=backup -e SFTP_IDENTITY_FILE=/config/id_ed25519 \
-e SFTP_KNOWN_HOSTS=/config/known_hosts \
jkaninda/pg-bkup backup --storage sftp -d shop --path /home/backup/shop

# FTP, Azure Blob, Google Cloud Storage
jkaninda/pg-bkup backup --storage ftp   -d shop --path /backups   # FTP_HOST, FTP_USER, FTP_PASSWORD [FTP_TLS]
jkaninda/pg-bkup backup --storage azure -d shop --path backups    # AZURE_STORAGE_ACCOUNT_NAME, _KEY (or _SAS_TOKEN), _CONTAINER_NAME
jkaninda/pg-bkup backup --storage gcs   -d shop --path backups    # GCS_BUCKET, GCS_CREDENTIALS_FILE
```

`restore` takes the same `--storage` and `--path` plus the bare file name: `restore --storage s3 --path /prod/shop -d shop -f shop_20261027_031500.sql.gz.gpg`.

## Encryption

| Mode | Backup | Restore |
|---|---|---|
| GPG passphrase | `GPG_PASSPHRASE` | `GPG_PASSPHRASE` |
| GPG public key | `GPG_PUBLIC_KEY` (path or inline), or `/config/gnupg/public_key.{asc,gpg}` | `GPG_PRIVATE_KEY` (+ `GPG_PRIVATE_KEY_PASSPHRASE`, defaults to `GPG_PASSPHRASE`) |
| age recipients | `AGE_RECIPIENTS` (`age1...`, `ssh-ed25519 ...`) or `AGE_RECIPIENTS_FILE` | `AGE_IDENTITY` / `AGE_IDENTITY_FILE` (default `/config/age/identity.txt`) |
| age passphrase | `AGE_PASSPHRASE` | `AGE_PASSPHRASE` |

`ENCRYPTION=auto` (default) picks from the keys that are set; GPG wins when both are set. Encryption fails closed: a configured but unusable key aborts the backup. Encrypted files end in `.gpg` or `.age`.

## Backup index and retention

Each storage directory holds an index (`BACKUP_INDEX_NAME`, default `pg-bkup-metadata.json`, placeholders `{tool}` `{engine}` `{name}` `{db}`) and a `<artifact>.meta.json` sidecar per backup, with size, sha256, server version and encryption.

- `BACKUP_RETENTION_DAYS=7` and/or `BACKUP_RETENTION_COUNT=10` prune after each backup; the newest backup is always kept.
- Only files recorded in the index are deleted. Backups made before the index existed are imported with `BACKUP_RETENTION_ADOPT_LEGACY=true`; `BACKUP_INDEX=false` falls back to name-pattern retention.
- `list`, `verify -f FILE`, `delete -f FILE`, `prune [--dry-run]` and `reconcile` work from the index.

## Scheduled backups

```shell
docker run -d --name pg-bkup --network your_network -v $PWD/backup:/backup \
  -e DB_HOST=postgres -e DB_USER=postgres -e DB_PASSWORD=password -e BACKUP_RETENTION_DAYS=7 \
  jkaninda/pg-bkup backup -d database_name --cron-expression "@daily"
```

Cron expressions take 5 or 6 fields, `@daily`/`@hourly`/... or `@every 6h`. `BACKUP_RUN_ON_START=true` runs one backup immediately. A failed run is logged and notified and the daemon keeps running. `docker inspect` shows the health status (`pg-bkup health`); `METRICS_ADDR=:9090` serves Prometheus metrics.

On Kubernetes, prefer a `CronJob` running a one-shot `backup` (see the [docs](https://jkaninda.github.io/pg-bkup/how-tos/deploy-on-kubernetes/)).

## Several databases in one run

```yaml
# /config/config.yaml (or --config / BACKUP_CONFIG_FILE)
cronExpression: "@daily"      # optional
databases:
  - host: lldap-db
    name: lldap
    user: lldap
    password: ${LLDAP_DB_PASSWORD}   # ${VAR} and ${VAR:-default} are expanded
    path: /s3-path/lldap
  - name: keycloak                    # host/user/password from DB_HOST_KEYCLOAK or KEYCLOAK_DB_HOST, ... then DB_HOST
    path: /s3-path/keycloak
    sslMode: require
```

For each entry: the YAML field wins, then `<KEY>_<NAME>` / `<NAME>_<KEY>` env vars (e.g. `DB_PASSWORD_KEYCLOAK`, `KEYCLOAK_DB_PASSWORD`), then the global variable.

## Exit codes

| Code | Meaning |
|---|---|
| 0 | success |
| 1 | failure (dump, restore, ...) |
| 2 | configuration or usage error |
| 3 | connection or authentication error |
| 4 | storage error |
| 5 | encryption error |

`--result-file PATH` writes the result as JSON; `LOG_FORMAT=json` makes the logs machine-readable.

## Commands

```
pg-bkup backup     [-d DB | -a | -A] [--storage S] [--path P] [--cron-expression E]
pg-bkup restore    -f FILE [-d DB] [--storage S] [--path P]
pg-bkup migrate    [-d DB | -a | -I]
pg-bkup verify     -f FILE           # download, decrypt, decompress, check sha256
pg-bkup list       [--json] [--storage-scan]
pg-bkup delete     -f FILE
pg-bkup prune      [--dry-run]
pg-bkup reconcile  [--from-index OLD] [--dry-run]
pg-bkup config     [--all] [--env-docs]
pg-bkup health
pg-bkup version
```

The image also provides `backup`, `restore` and `migrate` wrapper scripts and the `bkup` alias, so `docker run jkaninda/pg-bkup backup ...` and `docker exec pg-bkup bkup backup ...` both work.

## Upgrading from 1.x

Every 1.x variable and flag still works; renamed variables log a deprecation warning once. What changed:

- **Restores are atomic.** Database dumps are replayed with `ON_ERROR_STOP=1 --single-transaction`: the first error rolls everything back and the command fails (1.x continued after errors). v2 dumps include `DROP ... IF EXISTS` statements (`PG_CLEAN`, default on), so restoring one over an existing database replaces its objects inside that transaction. A 1.x dump (no drops) restored into a database that already contains its tables now fails instead of half-applying: restore it into an empty database, or set `RESTORE_SINGLE_TRANSACTION=false` and/or `RESTORE_ON_ERROR_STOP=false` for the 1.x behaviour.
- **`--schema-only` and `--tables` backups are compressed** (`shop_schema_<timestamp>.sql.gz`; 1.x wrote `.sql`). `--data-only` files get a `_data_` infix. Use `--disable-compression` for the old extension.
- **Client versions:** `pg_dump` 16, 17 or 18 is chosen per server; `PG_CLIENT_VERSION` forces one. Dumps of servers older than 17 are taken with `pg_dump` 16, so they restore on those servers.
- **`DB_SSL_MODE` is passed to `pg_dump`, `psql` and `pg_restore`** (1.x only used it for the connection test). The default is now libpq's own (`prefer`) instead of `disable`; `DB_SSL_CA`, `DB_SSL_CERT` and `DB_SSL_KEY` are new.
- **`DB_URL`** also accepts `postgres://` URIs (credentials percent-encoded). An explicit `-d`/`DB_NAME` now wins over the URL's database.
- **Migrate** is a direct pipe (no temp file) and honours `TARGET_DB_SSL_MODE` and `TARGET_DB_AUTH_DATABASE`, which 1.x ignored.
- **Retention** only deletes backups recorded in the index (see above). Set `BACKUP_RETENTION_ADOPT_LEGACY=true` once to adopt 1.x backups.
- **`STORAGE` is honoured** (1.x always used the flag's default `local` unless `--storage` was given). Unknown storage names are errors instead of falling back to local.
- **SFTP variables are renamed, and the old names are no longer read:** `SSH_HOST`/`SSH_HOST_NAME` → `SFTP_HOST`, `SSH_PORT` → `SFTP_PORT`, `SSH_USER` → `SFTP_USER`, `SSH_PASSWORD` → `SFTP_PASSWORD`, `SSH_IDENTIFY_FILE`/`SSH_IDENTITY_FILE` → `SFTP_IDENTITY_FILE`, `SSH_REMOTE_PATH` → `REMOTE_PATH`.
- **SFTP host keys are verified.** Set `SFTP_KNOWN_HOSTS`, or `SFTP_INSECURE_IGNORE_HOST_KEY=true` for the 1.x behaviour. `SFTP_PORT` defaults to 22.
- **`MAIL_SKIP_TLS` now means what it says** (`true` skips certificate verification); 1.x inverted it. Prefer `MAIL_INSECURE_SKIP_VERIFY` and `MAIL_TLS=starttls|tls|none`.
- `AWS_REGION` defaults to `us-east-1`; `AWS_ACCESS_KEY`/`AWS_SECRET_KEY` may be omitted to use the AWS credential chain.
- The `templates` directory is built in; custom notification templates go in `TEMPLATES_DIR` (default `/config/templates`).
- The image runs `tini` as PID 1 and has a `HEALTHCHECK` (daemon mode only). Tags: `latest`, `vX.Y.Z`, `vX.Y`, `vX`; pin a version in production.

## Environment variables

Generated with `pg-bkup config --env-docs`. Every secret also accepts `<NAME>_FILE` (Docker/Kubernetes secrets). Resolution order: flag > variable > deprecated name > config file > default.

<!-- env-docs:start -->
### storage

| Variable | Flag | Default | Fallbacks | Description |
|---|---|---|---|---|
| `STORAGE` | `--storage` `-s` | `local` |  | Storage backend: local, s3, gcs, azure, sftp (ssh), ftp (local, s3, gcs, azure, sftp, ssh, ftp, remote) |
| `STORAGE_PATH` |  | `/backup` |  | Root directory of the local storage backend |
| `REMOTE_PATH` | `--path` `-P` |  | `AWS_S3_PATH` (deprecated), `S3_PATH` (deprecated) | Folder inside the storage backend (without file name) |

### general

| Variable | Flag | Default | Fallbacks | Description |
|---|---|---|---|---|
| `BACKUP_NAME` | `--name` `-n` |  |  | Base name of the artifact; a timestamp is appended (default: the database or volume name) |
| `BACKUP_CUSTOM_NAME` | `--custom-name` |  |  | Fixed artifact name without timestamp (overwrites the previous one) |
| `FILE_NAME` | `--file` `-f` |  |  | Backup file to restore, verify or delete |
| `BACKUP_CONFIG_FILE` | `--config` `-c` |  |  | YAML configuration file (default /config/config.yaml if present) |
| `BACKUP_TIMEOUT` |  |  |  | Abort a backup after this duration (0 = none) |
| `RESTORE_TIMEOUT` |  |  |  | Abort a restore after this duration (0 = none) |
| `TMP_DIR` |  | `/tmp/backup` | `BACKUP_TMP_DIR` (deprecated) | Scratch directory (only for secret files and disk-based modes) |
| `COMPRESSION` | `--compression` | `gzip` |  | Compression: gzip, zstd or none (gzip, zstd, none) |
| `COMPRESSION_LEVEL` |  |  |  | Compression level (0 = default) |
| `DISABLE_COMPRESSION` | `--disable-compression` |  |  | Same as COMPRESSION=none |
| `RESTORE_VERIFY_FIRST` | `--verify-first` | `auto` |  | Verify the whole artifact before restoring (auto: on for GPG-encrypted volumes) (auto, true, false) |
| `LOG_LEVEL` |  | `info` |  | Log level (debug, info, warn, error) |
| `LOG_FORMAT` |  | `text` |  | Log format (text, json) |
| `TIME_FORMAT` |  | `2006-01-02 15:04:05` |  | Time layout used in notifications |
| `RESULT_FILE` | `--result-file` |  |  | Write the run result as JSON to this path |

### scheduling

| Variable | Flag | Default | Fallbacks | Description |
|---|---|---|---|---|
| `BACKUP_CRON_EXPRESSION` | `--cron-expression` `-e` |  |  | Run as a daemon on this cron schedule (5 or 6 fields, @daily, @every 6h) |
| `BACKUP_RUN_ON_START` |  |  |  | In daemon mode, run one backup immediately at start |
| `BACKUP_RESCUE_MODE` |  |  |  | Deprecated: runs always continue with the remaining targets after a failure |
| `METRICS_ADDR` |  |  |  | Serve Prometheus /metrics and /healthz on this address in daemon mode (e.g. :9090) |
| `HEALTH_FILE` |  | `/tmp/bkup-health.json` |  | State file read by the health command |
| `HEALTH_MAX_AGE` |  |  |  | health fails when the last success is older than this (default: 2x the schedule interval) |

### retention

| Variable | Flag | Default | Fallbacks | Description |
|---|---|---|---|---|
| `BACKUP_RETENTION_DAYS` |  | `0` |  | Delete backups older than N days (0 = keep) |
| `BACKUP_RETENTION_COUNT` |  | `0` |  | Keep only the newest N backups (0 = no limit) |
| `BACKUP_RETENTION_ALLOW_ROOT` |  |  |  | Allow pattern-based retention at the storage root (index disabled) |
| `BACKUP_RETENTION_ADOPT_LEGACY` |  |  |  | Import pre-index backups matching this tool's naming into the index |
| `BACKUP_INDEX` |  | `true` |  | Keep the backup index file and per-backup .meta.json sidecars |
| `BACKUP_INDEX_NAME` | `--index-name` | `{tool}-metadata.json` |  | Index file name; placeholders {tool} {engine} {name} {db} |

### storage-s3

| Variable | Flag | Default | Fallbacks | Description |
|---|---|---|---|---|
| `AWS_S3_ENDPOINT` |  |  | `S3_ENDPOINT` (deprecated), `AWS_ENDPOINT_URL_S3`, `AWS_ENDPOINT_URL` | S3 endpoint (empty = AWS) |
| `AWS_S3_BUCKET_NAME` |  |  | `BUCKET_NAME` (deprecated), `AWS_BUCKET_NAME` (deprecated) | S3 bucket |
| `AWS_REGION` |  | `us-east-1` | `AWS_DEFAULT_REGION` | S3 region |
| `AWS_ACCESS_KEY` |  |  | `ACCESS_KEY` (deprecated), `AWS_ACCESS_KEY_ID` | S3 access key (empty = default AWS credential chain) |
| `AWS_SECRET_KEY` |  |  | `SECRET_KEY` (deprecated), `AWS_SECRET_ACCESS_KEY`, `AWS_SECRET_KEY_FILE` | S3 secret key |
| `AWS_SESSION_TOKEN` |  |  | `AWS_SESSION_TOKEN_FILE` | S3 session token |
| `AWS_PROFILE` |  |  |  | AWS shared config profile |
| `AWS_DISABLE_SSL` |  |  |  | Use http for the S3 endpoint |
| `AWS_FORCE_PATH_STYLE` |  |  |  | Path-style S3 addressing (MinIO) |
| `S3_STORAGE_CLASS` |  |  |  | S3 storage class |
| `S3_SSE` |  |  |  | Server-side encryption: AES256 or aws:kms (aes256, aws:kms) |
| `S3_SSE_KMS_KEY_ID` |  |  |  | KMS key id for aws:kms |
| `S3_PART_SIZE` |  | `64MiB` |  | Multipart part size; max object size is 10000 x this |
| `S3_UPLOAD_CONCURRENCY` |  | `2` |  | Parallel part uploads (memory = part size x this) |

### storage-sftp

| Variable | Flag | Default | Fallbacks | Description |
|---|---|---|---|---|
| `SFTP_HOST` |  |  |  | SFTP host |
| `SFTP_PORT` |  | `22` |  | SFTP port |
| `SFTP_USER` |  |  |  | SFTP user |
| `SFTP_PASSWORD` |  |  | `SFTP_PASSWORD_FILE` | SFTP password |
| `SFTP_IDENTITY_FILE` |  |  |  | SFTP private key file (the SSH agent is also used when SSH_AUTH_SOCK is set) |
| `SFTP_IDENTITY_PASSPHRASE` |  |  | `SFTP_IDENTITY_PASSPHRASE_FILE` | Passphrase of the SFTP private key |
| `SFTP_KNOWN_HOSTS` |  |  |  | known_hosts file used to verify the SFTP host key (default ~/.ssh/known_hosts) |
| `SFTP_INSECURE_IGNORE_HOST_KEY` |  |  |  | Do not verify the SFTP host key (insecure) |

### storage-ftp

| Variable | Flag | Default | Fallbacks | Description |
|---|---|---|---|---|
| `FTP_HOST` |  |  | `FTP_HOST_NAME` (deprecated) | FTP host |
| `FTP_PORT` |  | `21` |  | FTP port |
| `FTP_USER` |  |  |  | FTP user |
| `FTP_PASSWORD` |  |  | `FTP_PASSWORD_FILE` | FTP password |
| `FTP_TLS` |  |  |  | Use explicit FTPS |

### storage-azure

| Variable | Flag | Default | Fallbacks | Description |
|---|---|---|---|---|
| `AZURE_STORAGE_ACCOUNT_NAME` |  |  |  | Azure storage account |
| `AZURE_STORAGE_ACCOUNT_KEY` |  |  | `AZURE_STORAGE_ACCOUNT_KEY_FILE` | Azure storage account key |
| `AZURE_STORAGE_CONTAINER_NAME` |  |  |  | Azure blob container |
| `AZURE_STORAGE_ENDPOINT` |  |  |  | Azure blob endpoint (Azurite, sovereign clouds) |
| `AZURE_STORAGE_SAS_TOKEN` |  |  | `AZURE_STORAGE_SAS_TOKEN_FILE` | Azure SAS token (instead of the account key) |

### storage-gcs

| Variable | Flag | Default | Fallbacks | Description |
|---|---|---|---|---|
| `GCS_BUCKET` |  |  |  | GCS bucket |
| `GCS_CREDENTIALS_FILE` |  |  | `GOOGLE_APPLICATION_CREDENTIALS` | GCS service account JSON file |
| `GCS_CREDENTIALS_JSON` |  |  | `GCS_CREDENTIALS_JSON_FILE` | GCS service account JSON content |
| `GCS_ENDPOINT` |  |  |  | GCS endpoint (emulators) |

### encryption

| Variable | Flag | Default | Fallbacks | Description |
|---|---|---|---|---|
| `ENCRYPTION` | `--encryption` | `auto` |  | auto (from the keys set), none, gpg or age (auto, none, gpg, age) |
| `ENCRYPTION_ARMOR` |  |  |  | ASCII-armor encrypted output |
| `GPG_PASSPHRASE` |  |  | `GPG_PASSPHRASE_FILE` | GPG symmetric passphrase |
| `GPG_PUBLIC_KEY` |  |  |  | GPG public key(s): file path or inline key (default /config/gnupg/public_key.{asc,gpg}) |
| `GPG_PRIVATE_KEY` |  |  | `GPG_PRIVATE_KEY_FILE` | GPG private key for restore: file path or inline key (default /config/gnupg/private_key.{asc,gpg}) |
| `GPG_PRIVATE_KEY_PASSPHRASE` |  |  | `GPG_PRIVATE_KEY_PASSPHRASE_FILE` | Passphrase of the GPG private key (default GPG_PASSPHRASE) |
| `AGE_RECIPIENTS` |  |  |  | age recipients (age1..., ssh-ed25519 ...), comma or newline separated |
| `AGE_RECIPIENTS_FILE` |  |  |  | File with age recipients |
| `AGE_PASSPHRASE` |  |  | `AGE_PASSPHRASE_FILE` | age passphrase (scrypt) |
| `AGE_IDENTITY` |  |  | `AGE_IDENTITY_FILE` | age identity (AGE-SECRET-KEY-...) for restore |
| `AGE_IDENTITY_FILE` |  |  |  | age identity file for restore (default /config/age/identity.txt) |

### notifications

| Variable | Flag | Default | Fallbacks | Description |
|---|---|---|---|---|
| `MAIL_HOST` |  |  |  | SMTP host |
| `MAIL_PORT` |  | `587` |  | SMTP port |
| `MAIL_USERNAME` |  |  |  | SMTP user |
| `MAIL_PASSWORD` |  |  | `MAIL_PASSWORD_FILE` | SMTP password |
| `MAIL_FROM` |  |  |  | Sender address |
| `MAIL_TO` |  |  |  | Recipients, comma separated |
| `MAIL_SKIP_TLS` |  |  |  | Deprecated: true skips certificate verification (use MAIL_INSECURE_SKIP_VERIFY) |
| `MAIL_INSECURE_SKIP_VERIFY` |  |  |  | Skip SMTP certificate verification |
| `MAIL_TLS` |  | `starttls` |  | SMTP TLS mode (starttls, tls, none) |
| `TG_TOKEN` |  |  | `TELEGRAM_BOT_TOKEN`, `TG_TOKEN_FILE` | Telegram bot token |
| `TG_CHAT_ID` |  |  | `TELEGRAM_CHAT_ID` | Telegram chat id |
| `POSTA_URL` |  |  |  | Posta base URL (https://posta.example.com) for email notifications through Posta |
| `POSTA_API_KEY` |  |  | `POSTA_API_KEY_FILE` | Posta API key with the send scope |
| `POSTA_WORKSPACE_ID` |  |  |  | Posta workspace id (only for account-wide API keys) |
| `POSTA_FROM` |  |  |  | Posta sender address (default MAIL_FROM) |
| `POSTA_TO` |  |  |  | Posta recipients, comma separated (default MAIL_TO) |
| `POSTA_TEMPLATE` |  |  |  | Posta template name for notifications; the event fields are passed as template data (default: built-in HTML) |
| `POSTA_TEMPLATE_ERROR` |  |  |  | Posta template name for failures (default POSTA_TEMPLATE) |
| `POSTA_LANGUAGE` |  |  |  | Posta template language |
| `WEBHOOK_URL` |  |  |  | Generic JSON webhook URL |
| `WEBHOOK_SECRET` |  |  | `WEBHOOK_SECRET_FILE` | HMAC-SHA256 secret for the X-Bkup-Signature header |
| `SLACK_WEBHOOK_URL` |  |  | `SLACK_WEBHOOK_URL_FILE` | Slack incoming webhook URL |
| `DISCORD_WEBHOOK_URL` |  |  | `DISCORD_WEBHOOK_URL_FILE` | Discord webhook URL |
| `NTFY_URL` |  |  |  | ntfy topic URL |
| `NTFY_TOKEN` |  |  | `NTFY_TOKEN_FILE` | ntfy access token |
| `NOTIFY_ON` |  | `always` |  | When to notify (always, failure, success, never) |
| `HEALTHCHECK_PING_URL` |  |  |  | healthchecks.io-style URL pinged on start (/start), success and failure (/fail) |
| `BACKUP_REFERENCE` |  |  |  | Free text included in notifications |
| `TEMPLATES_DIR` |  | `/config/templates` |  | Directory with custom notification templates |

### database

| Variable | Flag | Default | Fallbacks | Description |
|---|---|---|---|---|
| `DB_URL` |  |  | `DB_URL_FILE` | Connection URL: postgres://user:pass@host:5432/db?sslmode=require or jdbc:postgresql://host:5432/db?user=u&password=p (overrides host, port, user, password; its database is used when DB_NAME is not set) |
| `DB_HOST` |  | `localhost` |  | Database host |
| `DB_PORT` |  | `5432` |  | Database port |
| `DB_NAME` | `--dbname` `-d` |  |  | Database name |
| `DB_USER` |  |  | `DB_USERNAME` | Database user |
| `DB_PASSWORD` |  |  | `DB_PASSWORD_FILE` | Database password |
| `DB_SSL_MODE` |  |  |  | libpq sslmode, passed to pg_dump, psql and pg_restore as PGSSLMODE (empty = libpq default, prefer) (disable, allow, prefer, require, verify-ca, verify-full) |
| `DB_SSL_CA` |  |  |  | CA certificate file (PGSSLROOTCERT) |
| `DB_SSL_CERT` |  |  |  | Client certificate file (PGSSLCERT) |
| `DB_SSL_KEY` |  |  |  | Client key file (PGSSLKEY) |
| `BACKUP_ALL_DATABASES` | `--all-databases` `-a` |  |  | backup: every database (templates and 'postgres' excluded), one file each; migrate: every database into the same name on the target |
| `BACKUP_ALL_IN_ONE` | `--all-in-one` `-A` |  |  | Back up the whole instance (roles, tablespaces, all databases) into one file with pg_dumpall |
| `BACKUP_EXCLUDE_DATABASES` | `--exclude-db` |  |  | Databases to skip with --all-databases / --all-in-one, comma separated (repeatable flag) |
| `TARGET_DB_URL` |  |  | `TARGET_DB_URL_FILE` | migrate: target connection URL (postgres:// or jdbc:postgresql://) |
| `TARGET_DB_HOST` |  |  |  | migrate: target host |
| `TARGET_DB_PORT` |  | `5432` |  | migrate: target port |
| `TARGET_DB_NAME` |  |  |  | migrate: target database |
| `TARGET_DB_USER` |  |  | `TARGET_DB_USERNAME` | migrate: target user |
| `TARGET_DB_PASSWORD` |  |  | `TARGET_DB_PASSWORD_FILE` | migrate: target password |
| `TARGET_DB_SSL_MODE` |  |  |  | migrate: target libpq sslmode (disable, allow, prefer, require, verify-ca, verify-full) |
| `TARGET_DB_SSL_CA` |  |  |  | migrate: target CA certificate file |
| `DB_AUTH_DATABASE` |  | `postgres` |  | Maintenance database used to list databases, create databases and run pg_dumpall |
| `TARGET_DB_AUTH_DATABASE` |  | `postgres` |  | migrate: target maintenance database |
| `TARGET_DB_SSL_CERT` |  |  |  | migrate: target client certificate file |
| `TARGET_DB_SSL_KEY` |  |  |  | migrate: target client key file |
| `PG_DUMP_FORMAT` | `--dump-format` | `plain` |  | Dump format: plain (.sql, psql) or custom (.dump, pg_dump -Fc -Z0, pg_restore); COMPRESSION applies to both (plain, custom) |
| `PG_CLIENT_VERSION` |  |  |  | Force a PostgreSQL client major (16, 17, 18); default: the lowest bundled client >= the server version |
| `BACKUP_SCHEMA_ONLY` | `--schema-only` |  |  | Dump the schema only (file name infix _schema_) |
| `BACKUP_DATA_ONLY` | `--data-only` |  |  | Dump the data only (file name infix _data_) |
| `BACKUP_TABLES` | `--tables` `-t` |  |  | Only these tables (pg_dump -t patterns), comma separated or repeated (file name infix _tables_N_) |
| `BACKUP_EXCLUDE_TABLES` | `--exclude-table` |  |  | Skip these tables (pg_dump --exclude-table patterns), comma separated or repeated |
| `BACKUP_NO_ROLE_PASSWORDS` | `--no-role-passwords` |  |  | --all-in-one / migrate --entire-instance: do not dump role passwords (managed services such as Aiven, OVH) |
| `MIGRATE_ENTIRE_INSTANCE` | `--entire-instance` `-I` |  |  | migrate: copy the whole instance (roles, tablespaces, all databases) with pg_dumpall |
| `RESTORE_SINGLE_TRANSACTION` |  | `true` |  | Restore a database dump in one transaction (all or nothing); not used for --all-in-one dumps |
| `RESTORE_ON_ERROR_STOP` |  | `auto` |  | Stop a psql restore at the first error; auto = on for database dumps, off for --all-in-one (pg_dumpall) dumps, which always contain 'role already exists' errors (auto, true, false) |
| `PG_CLEAN` |  | `true` |  | Drop objects before recreating them (pg_dump/pg_restore --clean --if-exists), so a restore replaces an existing database inside its transaction |
| `RESTORE_NO_OWNER` |  | `true` |  | pg_restore (custom format): do not restore object ownership (--no-owner) |

<!-- env-docs:end -->

## Available image registries

```
docker pull jkaninda/pg-bkup
docker pull ghcr.io/jkaninda/pg-bkup
```

## License

This project is licensed under the MIT License. See the LICENSE file for details.

## Authors

**Jonas Kaninda** - <https://github.com/jkaninda>

## Copyright

Copyright (c) [2024] [Jonas Kaninda]
