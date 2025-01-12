---
title: Backup
layout: default
parent: How Tos
nav_order: 1
---

# Backup Database

To back up your database, use the `backup` command. 

This section explains how to configure and run backups, including recurring backups, using Docker or Kubernetes.

---

## Default Configuration

- **Storage**: By default, backups are stored locally in the `/backup` directory.
- **Compression**: Backups are compressed using `gzip` by default. Use the `--disable-compression` flag to disable compression.
- **Security**: It is recommended to create a dedicated user with read-only access for backup tasks.

{: .note }
The backup process supports recurring backups on Docker or Docker Swarm. On Kubernetes, it can be deployed as a CronJob.

---

## Example: Basic Backup Configuration

Below is an example `docker-compose.yml` configuration for backing up a database:

```yaml
services:
  pg-bkup:
    # In production, lock your image tag to a specific release version
    # instead of using `latest`. Check https://github.com/jkaninda/pg-bkup/releases
    # for available releases.
    image: jkaninda/pg-bkup
    container_name: pg-bkup
    command: backup -d database
    volumes:
      - ./backup:/backup
    environment:
      - DB_PORT=5432
      - DB_HOST=postgres
      - DB_NAME=database
      - DB_USERNAME=username
      - DB_PASSWORD=password
      ## Optional: Use JDBC connection string
      #- DB_URL=jdbc:postgresql://postgres:5432/database?user=user&password=password

    # Ensure the pg-bkup container is connected to the same network as your database
    networks:
      - web

networks:
  web:
```

---

## Backup Using Docker CLI

You can also run backups directly using the Docker CLI:

```bash
docker run --rm --network your_network_name \
  -v $PWD/backup:/backup/ \
  -e "DB_HOST=dbhost" \
  -e "DB_USERNAME=username" \
  -e "DB_PASSWORD=password" \
  jkaninda/pg-bkup backup -d database_name
```

---

## Recurring Backups

To schedule recurring backups, use the `--cron-expression` flag or the `BACKUP_CRON_EXPRESSION` environment variable. This allows you to define a cron schedule for automated backups.

### Example: Recurring Backup Configuration

```yaml
services:
  pg-bkup:
    # In production, lock your image tag to a specific release version
    # instead of using `latest`. Check https://github.com/jkaninda/pg-bkup/releases
    # for available releases.
    image: jkaninda/pg-bkup
    container_name: pg-bkup
    command: backup -d database --cron-expression @midnight
    volumes:
      - ./backup:/backup
    environment:
      - DB_PORT=5432
      - DB_HOST=postgres
      - DB_NAME=database
      - DB_USERNAME=username
      - DB_PASSWORD=password
      ## Optional: Define a cron schedule for recurring backups
      - BACKUP_CRON_EXPRESSION=@midnight
      ## Optional: Delete old backups after a specified number of days
      #- BACKUP_RETENTION_DAYS=7
      ## Optional: Use JDBC connection string
      #- DB_URL=jdbc:postgresql://postgres:5432/database?user=user&password=password

    # Ensure the pg-bkup container is connected to the same network as your database
    networks:
      - web

networks:
  web:
```

---

## Key Notes

- **Cron Expression**: Use the `--cron-expression` flag or `BACKUP_CRON_EXPRESSION` environment variable to define the backup schedule. For example:
    - `@midnight`: Runs the backup daily at midnight.
    - `0 1 * * *`: Runs the backup daily at 1:00 AM.
- **Backup Retention**: Optionally, use the `BACKUP_RETENTION_DAYS` environment variable to automatically delete backups older than a specified number of days.
- **JDBC Connection**: You can use the `DB_URL` environment variable to specify a JDBC connection string instead of individual database credentials.
