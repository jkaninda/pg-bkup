---
title: Configuration Reference
layout: default
nav_order: 3
---

# Configuration Reference

pg-bkup is configured with **environment variables**, **CLI flags** and an optional **YAML file**.
For every setting the order is: flag (when given) > variable > deprecated name > config file > default.
Every secret also accepts `<NAME>_FILE` (Docker and Kubernetes secrets). `pg-bkup config` prints the resolved configuration with its sources (secrets redacted).

{: .note }
Upgrading from 1.x? All 1.x variables and flags still work. See [Upgrading to v2](../upgrade/) for the behaviour changes.

---

## Commands

| Command | Description |
|---|---|
| `backup` | Back up now, or as a daemon with `--cron-expression`. |
| `restore -f FILE` | Stream a backup back into `-d DB` (`.sql`, `.sql.gz`, `.sql.zst`, `.dump`, optionally `.gpg`/`.age`). |
| `migrate` | Copy `DB_*` into `TARGET_DB_*` with `pg_dump \| psql`, no file. |
| `verify -f FILE` | Download, decrypt and decompress a backup and check its sha256. |
| `list` | Backup history from the index (`--json`, `--all`, `--storage-scan`). |
| `delete -f FILE` | Delete a backup, its sidecar and its index entry. |
| `prune` | Apply the retention policy (`--dry-run`). |
| `reconcile` | Rebuild the index from the storage (`--from-index OLD` after a rename). |
| `config` | Show the resolved configuration; `--env-docs` prints the table below. |
| `health` | Docker `HEALTHCHECK` for daemon mode. |
| `version` | Print the version. |

The image also has `backup`, `restore` and `migrate` wrapper scripts and the `bkup` alias.

## Flags

| Flag | Short | Commands | Description |
|---|---|---|---|
| `--dbname` | `-d` | all | Database name (`DB_NAME`). |
| `--storage` | `-s` | all | `local`, `s3`, `gcs`, `azure`, `sftp` (`ssh`, `remote`), `ftp` (`STORAGE`). |
| `--path` | `-P` | all | Folder inside the storage (`REMOTE_PATH`); `--remote-path` is accepted. |
| `--file` | `-f` | restore, verify, delete | Backup file name. On local storage an absolute path such as `/backup/x.sql.gz` also works. |
| `--config` | `-c` | backup, list, prune, config | YAML configuration file (`BACKUP_CONFIG_FILE`, default `/config/config.yaml` if present). |
| `--cron-expression` | `-e` | backup | Run as a daemon on this schedule. |
| `--name` | `-n` | backup, list, prune | Base name of the artifact (default: the database name). |
| `--custom-name` | | backup | Fixed artifact name without timestamp. |
| `--all-databases` | `-a` | backup, migrate | Every database (templates and `postgres` excluded), one file each. |
| `--all-in-one` | `-A` | backup | The whole instance in one file with `pg_dumpall`. |
| `--exclude-db` | | backup | Databases to skip with `-a`/`-A` (comma separated or repeated). |
| `--no-role-passwords` | | backup, migrate | `pg_dumpall --no-role-passwords` (managed services such as Aiven, OVH, RDS). |
| `--schema-only` | | backup | Schema only (`<db>_schema_<time>.sql.gz`). |
| `--data-only` | | backup | Data only (`<db>_data_<time>.sql.gz`). |
| `--tables` | `-t` | backup | Only these tables, comma separated or repeated (`<db>_tables_<n>_<time>.sql.gz`). |
| `--exclude-table` | | backup | Skip these tables, comma separated or repeated. |
| `--dump-format` | | backup | `plain` (default) or `custom` (`PG_DUMP_FORMAT`). |
| `--compression` | | backup | `gzip` (default), `zstd` or `none`. |
| `--disable-compression` | | backup | Same as `--compression none`. |
| `--encryption` | | backup | `auto` (default), `none`, `gpg` or `age`. |
| `--entire-instance` | `-I` | migrate | Copy roles, tablespaces and every database with `pg_dumpall`. |
| `--index-name` | | all | Index file name (`BACKUP_INDEX_NAME`). |
| `--result-file` | | all | Write the run result as JSON. |
| `--verify-first` | | restore | Verify the whole artifact before restoring. |

## Exit codes

`0` success, `1` failure, `2` configuration error, `3` connection or authentication error, `4` storage error, `5` encryption error.

## PostgreSQL clients and versions

The image contains the PostgreSQL 16, 17 and 18 clients (`/usr/libexec/postgresqlNN/`). For each server pg-bkup reads `server_version_num` and runs the lowest `pg_dump` that is not older than the server (servers 9.2 to 16 use `pg_dump` 16), so dumps can be replayed on the same server version. Restores use the newest `psql`/`pg_restore`. `PG_CLIENT_VERSION` forces a major.

Connection settings reach the clients through their environment (`PGHOST`, `PGPORT`, `PGUSER`, `PGPASSWORD`, `PGSSLMODE`, `PGSSLROOTCERT`, `PGSSLCERT`, `PGSSLKEY`), never through the command line.

## Restore behaviour

| Backup | Restored with |
|---|---|
| plain dump of one database | `psql -v ON_ERROR_STOP=1 --single-transaction -d DB` |
| custom format (`PGDMP` header, `.dump`) | `pg_restore --single-transaction --exit-on-error --no-owner -d DB` |
| `--all-in-one` (`pg_dumpall`) | `psql -d DB_AUTH_DATABASE` (or `-d`), no transaction, no `ON_ERROR_STOP` |

The format is detected from the first bytes of the decompressed stream. A failing download, decryption or decompression kills the client before it sees the end of its input, so nothing is committed. `RESTORE_SINGLE_TRANSACTION`, `RESTORE_ON_ERROR_STOP` and `RESTORE_NO_OWNER` change these defaults.

---

## Environment variables

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


---

## Scheduled Backups

### Running in Scheduled Mode

- **Docker**: Use the `--cron-expression` flag or the `BACKUP_CRON_EXPRESSION` environment variable to schedule backups.
- **Kubernetes**: Use a `CronJob` resource for scheduled backups.

### Cron Syntax

The cron syntax consists of five fields:

```conf
* * * * * command
```

| Field         | Description                  | Values         |
|---------------|------------------------------|----------------|
| Minute        | Minute of the hour           | `0-59`         |
| Hour          | Hour of the day              | `0-23`         |
| Day of Month  | Day of the month             | `1-31`         |
| Month         | Month of the year            | `1-12`         |
| Day of Week   | Day of the week (0 = Sunday) | `0-7`          |

#### Examples

- **Every 30 minutes**: `*/30 * * * *`
- **Every hour at minute 0**: `0 * * * *`
- **Every day at 1:00 AM**: `0 1 * * *`

### Predefined Schedules

| Entry                      | Description                                | Equivalent To |
|----------------------------|--------------------------------------------|---------------|
| `@yearly` (or `@annually`) | Run once a year, midnight, Jan. 1st        | `0 0 1 1 *`   |
| `@monthly`                 | Run once a month, midnight, first of month | `0 0 1 * *`   |
| `@weekly`                  | Run once a week, midnight between Sat/Sun  | `0 0 * * 0`   |
| `@daily` (or `@midnight`)  | Run once a day, midnight                   | `0 0 * * *`   |
| `@hourly`                  | Run once an hour, beginning of hour        | `0 * * * *`   |

### Intervals

You can also schedule backups at fixed intervals using the format:

```conf
@every <duration>
```

- Example: `@every 1h30m10s` runs the backup every 1 hour, 30 minutes, and 10 seconds.

v2 also accepts 6-field expressions (with seconds). In daemon mode a failed run is logged and notified, and the daemon keeps running; `BACKUP_RUN_ON_START=true` runs one backup at start.
