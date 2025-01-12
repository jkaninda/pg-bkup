---
title: Backup to AWS S3
layout: default
parent: How Tos
nav_order: 2
---
# Backup to AWS S3

To store your backups on AWS S3, you can configure the backup process to use the `--storage s3` option. This section explains how to set up and configure S3-based backups.

---

## Configuration Steps

1. **Specify the Storage Type**  
   Add the `--storage s3` flag to your backup command.

2. **Set the S3 Path**  
   Optionally, specify a custom folder within your S3 bucket where backups will be stored using the `--path` flag.  
   Example: `--path /my-custom-path`.

3. **Required Environment Variables**  
   The following environment variables are mandatory for S3-based backups:

    - `AWS_S3_ENDPOINT`: The S3 endpoint URL (e.g., `https://s3.amazonaws.com`).
    - `AWS_S3_BUCKET_NAME`: The name of the S3 bucket where backups will be stored.
    - `AWS_REGION`: The AWS region where the bucket is located (e.g., `us-west-2`).
    - `AWS_ACCESS_KEY`: Your AWS access key.
    - `AWS_SECRET_KEY`: Your AWS secret key.
    - `AWS_DISABLE_SSL`: Set to `"true"` if using an S3 alternative like Minio without SSL (default is `"false"`).
    - `AWS_FORCE_PATH_STYLE`: Set to `"true"` if using an S3 alternative like Minio (default is `"false"`).

---

## Example Configuration

Below is an example `docker-compose.yml` configuration for backing up to AWS S3:

```yaml
services:
  pg-bkup:
    # In production, lock your image tag to a specific release version
    # instead of using `latest`. Check https://github.com/jkaninda/pg-bkup/releases
    # for available releases.
    image: jkaninda/pg-bkup
    container_name: pg-bkup
    command: backup --storage s3 -d database --path /my-custom-path
    environment:
      - DB_PORT=5432
      - DB_HOST=postgres
      - DB_NAME=database
      - DB_USERNAME=username
      - DB_PASSWORD=password
      ## AWS Configuration
      - AWS_S3_ENDPOINT=https://s3.amazonaws.com
      - AWS_S3_BUCKET_NAME=backup
      - AWS_REGION=us-west-2
      - AWS_ACCESS_KEY=xxxx
      - AWS_SECRET_KEY=xxxxx
      ## Optional: Disable SSL for S3 alternatives like Minio
      - AWS_DISABLE_SSL="false"
      ## Optional: Enable path-style access for S3 alternatives like Minio
      - AWS_FORCE_PATH_STYLE=false

    # Ensure the pg-bkup container is connected to the same network as your database
    networks:
      - web

networks:
  web:
```

---

## Recurring Backups to S3

To schedule recurring backups to S3, use the `--cron-expression` flag or the `BACKUP_CRON_EXPRESSION` environment variable. This allows you to define a cron schedule for automated backups.

### Example: Recurring Backup Configuration

```yaml
services:
  pg-bkup:
    # In production, lock your image tag to a specific release version
    # instead of using `latest`. Check https://github.com/jkaninda/pg-bkup/releases
    # for available releases.
    image: jkaninda/pg-bkup
    container_name: pg-bkup
    command: backup --storage s3 -d database --cron-expression "0 1 * * *"
    environment:
      - DB_PORT=5432
      - DB_HOST=postgres
      - DB_NAME=database
      - DB_USERNAME=username
      - DB_PASSWORD=password
      ## AWS Configuration
      - AWS_S3_ENDPOINT=https://s3.amazonaws.com
      - AWS_S3_BUCKET_NAME=backup
      - AWS_REGION=us-west-2
      - AWS_ACCESS_KEY=xxxx
      - AWS_SECRET_KEY=xxxxx
      ## Optional: Define a cron schedule for recurring backups
      #- BACKUP_CRON_EXPRESSION=0 1 * * *
      ## Optional: Delete old backups after a specified number of days
      #- BACKUP_RETENTION_DAYS=7
      ## Optional: Disable SSL for S3 alternatives like Minio
      - AWS_DISABLE_SSL="false"
      ## Optional: Enable path-style access for S3 alternatives like Minio
      - AWS_FORCE_PATH_STYLE=false

    # Ensure the pg-bkup container is connected to the same network as your database
    networks:
      - web

networks:
  web:
```

---

## Key Notes

- **Cron Expression**: Use the `--cron-expression` flag or `BACKUP_CRON_EXPRESSION` environment variable to define the backup schedule. For example, `0 1 * * *` runs the backup daily at 1:00 AM.
- **Backup Retention**: Optionally, use the `BACKUP_RETENTION_DAYS` environment variable to automatically delete backups older than a specified number of days.
- **S3 Alternatives**: If using an S3 alternative like Minio, set `AWS_DISABLE_SSL="true"` and `AWS_FORCE_PATH_STYLE="true"` as needed.

