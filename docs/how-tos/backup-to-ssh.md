---
title: Backup to SSH or SFTP
layout: default
parent: How Tos
nav_order: 3
---
# Backup to SFTP or SSH Remote Server

To store your backups on an `SFTP` or `SSH` remote server instead of the default storage, you can configure the backup process to use the `--storage ssh` or `--storage remote` option.
This section explains how to set up and configure SSH-based backups.

---

## Configuration Steps

1. **Specify the Storage Type**  
   Add the `--storage ssh` or `--storage remote` flag to your backup command.

2. **Set the Remote Path**  
   Define the full remote path where backups will be stored using the `--path` flag or the `REMOTE_PATH` environment variable.  
   Example: `--path /home/jkaninda/backups`.

3. **Required Environment Variables**  
   The following environment variables are mandatory for SSH-based backups:

    - `SSH_HOST`: The hostname or IP address of the remote server.
    - `SSH_USER`: The username for SSH authentication.
    - `REMOTE_PATH`: The directory on the remote server where backups will be stored.
    - `SSH_IDENTIFY_FILE`: The path to the private key file for SSH authentication.
    - `SSH_PORT`: The SSH port (default is `22`).
    - `SSH_PASSWORD`: (Optional) Use this only if you are not using a private key for authentication.

   {: .note }
   **Security Recommendation**: Using a private key (`SSH_IDENTIFY_FILE`) is strongly recommended over password-based authentication (`SSH_PASSWORD`) for better security.

---

## Example Configuration

Below is an example `docker-compose.yml` configuration for backing up to an SSH remote server:

```yaml
services:
  pg-bkup:
    # In production, lock your image tag to a specific release version
    # instead of using `latest`. Check https://github.com/jkaninda/pg-bkup/releases
    # for available releases.
    image: jkaninda/pg-bkup
    container_name: pg-bkup
    command: backup --storage remote -d database
    volumes:
      - ./id_ed25519:/tmp/id_ed25519
    environment:
      - DB_PORT=5432
      - DB_HOST=postgres
      - DB_NAME=database
      - DB_USERNAME=username
      - DB_PASSWORD=password
      ## SSH Configuration
      - SSH_HOST="hostname"
      - SSH_PORT=22
      - SSH_USER=user
      - REMOTE_PATH=/home/jkaninda/backups
      - SSH_IDENTIFY_FILE=/tmp/id_ed25519
      ## Optional: Use password instead of private key (not recommended)
      #- SSH_PASSWORD=password

    # Ensure the pg-bkup container is connected to the same network as your database
    networks:
      - web

networks:
  web:
```

---

## Recurring Backups to SSH Remote Server

To schedule recurring backups, you can use the `--cron-expression` flag or the `BACKUP_CRON_EXPRESSION` environment variable. 
This allows you to define a cron schedule for automated backups.

### Example: Recurring Backup Configuration

```yaml
services:
  pg-bkup:
    # In production, lock your image tag to a specific release version
    # instead of using `latest`. Check https://github.com/jkaninda/pg-bkup/releases
    # for available releases.
    image: jkaninda/pg-bkup
    container_name: pg-bkup
    command: backup -d database --storage ssh --cron-expression "@daily"
    volumes:
      - ./id_ed25519:/tmp/id_ed25519
    environment:
      - DB_PORT=5432
      - DB_HOST=postgres
      - DB_NAME=database
      - DB_USERNAME=username
      - DB_PASSWORD=password
      ## SSH Configuration
      - SSH_HOST="hostname"
      - SSH_PORT=22
      - SSH_USER=user
      - REMOTE_PATH=/home/jkaninda/backups
      - SSH_IDENTIFY_FILE=/tmp/id_ed25519
      ## Optional: Delete old backups after a specified number of days
      #- BACKUP_RETENTION_DAYS=7
      ## Optional: Use password instead of private key (not recommended)
      #- SSH_PASSWORD=password

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
- **Security**: Always prefer private key authentication (`SSH_IDENTIFY_FILE`) over password-based authentication (`SSH_PASSWORD`) for enhanced security.

---