---
title: Upgrading to v2
layout: default
nav_order: 5
permalink: /upgrade/
---

# Upgrading from 1.x to v2

v2 is a rewrite on a shared core used by all the *-bkup tools. The command line, the image, the file names and every 1.x variable keep working, so most setups upgrade by changing the image tag. Read the behaviour changes below before upgrading production restores.

## What stays the same

- `backup`, `restore` and `migrate` with the same flags: `-d`, `-f`, `-s/--storage`, `-P/--path`, `-c/--config`, `-e/--cron-expression`, `-a/--all-databases`, `-A/--all-in-one`, `--exclude-db`, `--no-role-passwords`, `--schema-only`, `--data-only`, `-t/--tables`, `--custom-name`, `--disable-compression`, `migrate -a` and `migrate -I/--entire-instance`.
- File names: `<db>_<YYYYMMDD_HHMMSS>.sql.gz[.gpg]`, `_schema_`, `_tables_<n>_`, `all_databases_...`; GPG encryption with `GPG_PASSPHRASE` / `GPG_PUBLIC_KEY`, gzip by default.
- The last line of a successful backup names the artifact (`msg="Backup completed" ... file=<name>`).
- The image: `/backup`, `/config`, `/bin/sh`, the `backup`/`restore`/`migrate` scripts and the `bkup` alias.
- The multi-database YAML file (`databases:` with `host`, `port`, `name`, `user`, `password`, `path`, `sslMode`, `authDatabase`), `${VAR}` expansion and `DB_PASSWORD_<NAME>` / `<NAME>_DB_PASSWORD` variables.

## Renamed variables (old names still accepted)

Old names log a deprecation warning once per run.

`DB_USER` (and `TARGET_DB_USER`) is the database user variable; `DB_USERNAME` and `TARGET_DB_USERNAME` are still read without a warning, including the per-database forms (`DB_USERNAME_<NAME>`, `<NAME>_DB_USERNAME`).

| 1.x | v2 |
|---|---|
| `AWS_S3_PATH`, `S3_PATH` | `REMOTE_PATH` (or `--path`) |
| `S3_ENDPOINT` | `AWS_S3_ENDPOINT` |
| `BUCKET_NAME`, `AWS_BUCKET_NAME` | `AWS_S3_BUCKET_NAME` |
| `ACCESS_KEY`, `SECRET_KEY` | `AWS_ACCESS_KEY`, `AWS_SECRET_KEY` |
| `FTP_HOST_NAME` | `FTP_HOST` |
| `BACKUP_TMP_DIR` | `TMP_DIR` |
| `--storage remote` | `--storage sftp` (`ssh` and `remote` still work) |

New pg-bkup settings: `DB_SSL_CA`, `DB_SSL_CERT`, `DB_SSL_KEY`, `TARGET_DB_SSL_CERT`, `TARGET_DB_SSL_KEY`, `PG_DUMP_FORMAT` (`--dump-format`), `PG_CLIENT_VERSION`, `BACKUP_EXCLUDE_TABLES` (`--exclude-table`), `RESTORE_SINGLE_TRANSACTION`, `RESTORE_ON_ERROR_STOP`, `RESTORE_NO_OWNER`, `PG_CLEAN`. Every flag also has a variable (`BACKUP_SCHEMA_ONLY`, `BACKUP_TABLES`, `BACKUP_ALL_DATABASES`, ...): see the [reference](../reference/).

## Behaviour changes

- **Restores are atomic.** Database dumps are replayed with `ON_ERROR_STOP=1 --single-transaction`: the first error rolls everything back and the command fails (1.x continued after errors). v2 dumps include `DROP ... IF EXISTS` statements (`PG_CLEAN`, default on), so restoring one over an existing database replaces its objects inside that transaction. A 1.x dump (no drops) restored into a database that already contains its tables now fails instead of half-applying: restore it into an empty database, or set `RESTORE_SINGLE_TRANSACTION=false` and/or `RESTORE_ON_ERROR_STOP=false` for the 1.x behaviour. Statements that only fail because the server differs (ownership, grants to missing roles, unknown `SET` parameters) do not stop a restore: see [Restore behaviour](../reference/#restore-behaviour).
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
## Checklist

1. Pin the image: `jkaninda/pg-bkup:v2.x.y` (tags `vX.Y.Z`, `vX.Y`, `vX` and `latest`).
2. SFTP: rename the `SSH_*` variables to `SFTP_*` (see below; the old names are **no longer read**), mount a `known_hosts` file and set `SFTP_KNOWN_HOSTS`.
3. Email: replace `MAIL_SKIP_TLS` with `MAIL_INSECURE_SKIP_VERIFY` (check the value: the 1.x meaning was inverted).
4. Retention on existing backups: run once with `BACKUP_RETENTION_ADOPT_LEGACY=true`, or `pg-bkup reconcile`.
5. Restoring **1.x** dumps into non-empty databases: restore into an empty database, or set `RESTORE_SINGLE_TRANSACTION=false` and `RESTORE_ON_ERROR_STOP=false`. v2 dumps restore over existing databases (`PG_CLEAN`).
6. Run `pg-bkup config` to see every resolved value, its source and any deprecation warning.
