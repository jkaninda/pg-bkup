---
title: Configuration Reference
layout: default
nav_order: 2
---

# Configuration reference

Backup, restore and migrate targets, schedule and retention are configured using environment variables or flags.





###  CLI utility Usage

| Options               | Shorts | Usage                                                                                  |
|-----------------------|--------|----------------------------------------------------------------------------------------|
| pg-bkup               | bkup   | CLI utility                                                                            |
| backup                |        | Backup database operation                                                              |
| restore               |        | Restore database operation                                                             |
| migrate               |        | Migrate database from one instance to another one                                      |
| --storage             | -s     | Storage. local or s3 (default: local)                                                  |
| --file                | -f     | File name for restoration                                                              |
| --path                |        | AWS S3 path without file name. eg: /custom_path  or ssh remote path `/home/foo/backup` |
| --dbname              | -d     | Database name                                                                          |
| --port                | -p     | Database port (default: 5432)                                                          |
| --disable-compression |        | Disable database backup compression                                                    |
| --prune               |        | Delete old backup, default disabled                                                    |
| --keep-last           |        | Delete old backup created more than specified days ago, default 7 days                 |
| --cron-expression     |        | Backup cron expression, eg: (* * * * *) or @daily                                      |
| --help                | -h     | Print this help message and exit                                                       |
| --version             | -V     | Print version information and exit                                                     |

## Environment variables

| Name                   | Requirement                                                   | Description                                          |
|------------------------|---------------------------------------------------------------|------------------------------------------------------|
| DB_PORT                | Optional, default 5432                                        | Database port number                                 |
| DB_HOST                | Required                                                      | Database host                                        |
| DB_NAME                | Optional if it was provided from the -d flag                  | Database name                                        |
| DB_USERNAME            | Required                                                      | Database user name                                   |
| DB_PASSWORD            | Required                                                      | Database password                                    |
| AWS_ACCESS_KEY         | Optional, required for S3 storage                             | AWS S3 Access Key                                    |
| AWS_SECRET_KEY         | Optional, required for S3 storage                             | AWS S3 Secret Key                                    |
| AWS_BUCKET_NAME        | Optional, required for S3 storage                             | AWS S3 Bucket Name                                   |
| AWS_BUCKET_NAME        | Optional, required for S3 storage                             | AWS S3 Bucket Name                                   |
| AWS_REGION             | Optional, required for S3 storage                             | AWS Region                                           |
| AWS_DISABLE_SSL        | Optional, required for S3 storage                             | Disable SSL                                          |
| FILE_NAME              | Optional if it was provided from the --file flag              | Database file to restore (extensions: .sql, .sql.gz) |
| GPG_PASSPHRASE         | Optional, required to encrypt and restore backup              | GPG passphrase                                       |
| BACKUP_CRON_EXPRESSION | Optional if it was provided from the `--cron-expression` flag | Backup cron expression for docker in scheduled mode  |
| SSH_HOST_NAME          | Optional, required for SSH storage                            | ssh remote hostname or ip                            |
| SSH_USER               | Optional, required for SSH storage                            | ssh remote user                                      |
| SSH_PASSWORD           | Optional, required for SSH storage                            | ssh remote user's password                           |
| SSH_IDENTIFY_FILE      | Optional, required for SSH storage                            | ssh remote user's private key                        |
| SSH_PORT               | Optional, required for SSH storage                            | ssh remote server port                               |
| REMOTE_PATH            | Optional, required for SSH or FTP storage                     | remote path (/home/toto/backup)                      |
| FTP_HOST_NAME          | Optional, required for FTP storage                            | FTP host name                                        |
| FTP_PORT               | Optional, required for FTP storage                            | FTP server port number                               |
| FTP_USER               | Optional, required for FTP storage                            | FTP user                                             |
| FTP_PASSWORD           | Optional, required for FTP storage                            | FTP user password                                    |
| TARGET_DB_HOST         | Optional, required for database migration                     | Target database host                                 |
| TARGET_DB_PORT         | Optional, required for database migration                     | Target database port                                 |
| TARGET_DB_NAME         | Optional, required for database migration                     | Target database name                                 |
| TARGET_DB_USERNAME     | Optional, required for database migration                     | Target database username                             |
| TARGET_DB_PASSWORD     | Optional, required for database migration                     | Target database password                             |
| TG_TOKEN               | Optional, required for Telegram notification                  | Telegram token (`BOT-ID:BOT-TOKEN`)                  |
| TG_CHAT_ID             | Optional, required for Telegram notification                  | Telegram Chat ID                                     |
| TZ                     | Optional                                                      | Time Zone                                            |

---
## Run in Scheduled mode

This image can be run as CronJob in Kubernetes for a regular backup which makes deployment on Kubernetes easy as Kubernetes has CronJob resources.
For Docker, you need to run it in scheduled mode by adding `--cron-expression  "* * * * *"` flag or by defining `BACKUP_CRON_EXPRESSION=0 1 * * *` environment variable.

## Syntax of crontab (field description)

The syntax is:

- 1: Minute (0-59)
- 2: Hours (0-23)
- 3: Day (0-31)
- 4: Month (0-12 [12 == December])
- 5: Day of the week(0-7 [7 or 0 == sunday])

Easy to remember format:

```conf
* * * * * command to be executed
```

```conf
- - - - -
| | | | |
| | | | ----- Day of week (0 - 7) (Sunday=0 or 7)
| | | ------- Month (1 - 12)
| | --------- Day of month (1 - 31)
| ----------- Hour (0 - 23)
------------- Minute (0 - 59)
```

> At every 30th minute

```conf
*/30 * * * *
```
> “At minute 0.” every hour
```conf
0 * * * *
```

> “At 01:00.” every day

```conf
0 1 * * *
```
## Predefined schedules
You may use one of several pre-defined schedules in place of a cron expression.

| Entry                  | Description                                | Equivalent To |
|------------------------|--------------------------------------------|---------------|
| @yearly (or @annually) | Run once a year, midnight, Jan. 1st        | 0 0 1 1 *     |
| @monthly               | Run once a month, midnight, first of month | 0 0 1 * *     |
| @weekly                | Run once a week, midnight between Sat/Sun  | 0 0 * * 0     |
| @daily (or @midnight)  | Run once a day, midnight                   | 0 0 * * *     |
| @hourly                | Run once an hour, beginning of hour        | 0 * * * *     |

### Intervals
You may also schedule backup task at fixed intervals, starting at the time it's added or cron is run. This is supported by formatting the cron spec like this:

@every <duration>
where "duration" is a string accepted by time.

For example, "@every 1h30m10s" would indicate a schedule that activates after 1 hour, 30 minutes, 10 seconds, and then every interval after that.