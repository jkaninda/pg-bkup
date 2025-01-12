---
title: Restore database from SSH
layout: default
parent: How Tos
nav_order: 7
---

# Restore Database from SSH Remote Server

To restore a PostgreSQL database from a backup stored on an SSH remote server, use the `restore` command and specify the backup file with the `--file` flag. The system supports the following file formats:

- `.sql` (uncompressed SQL dump)
- `.sql.gz` (gzip-compressed SQL dump)
- `.sql.gpg` (GPG-encrypted SQL dump)
- `.sql.gz.gpg` (GPG-encrypted and gzip-compressed SQL dump)

---

## Configuration Steps

1. **Specify the Backup File**: Use the `--file` flag to specify the backup file to restore.
2. **Set the Storage Type**: Add the `--storage ssh` flag to indicate that the backup is stored on an SSH remote server.
3. **Provide SSH Configuration**: Include the necessary SSH credentials and configuration.
4. **Provide Database Credentials**: Ensure the correct database connection details are provided.

---

## Example: Restore from SSH Remote Server Configuration

Below is an example `docker-compose.yml` configuration for restoring a database from an SSH remote server:

```yaml
services:
  pg-bkup:
    # In production, lock your image tag to a specific release version
    # instead of using `latest`. Check https://github.com/jkaninda/pg-bkup/releases
    # for available releases.
    image: jkaninda/pg-bkup
    container_name: pg-bkup
    command: restore --storage ssh -d my-database -f store_20231219_022941.sql.gz --path /home/jkaninda/backups
    volumes:
      - ./backup:/backup  # Mount the directory for local operations (if needed)
      - ./id_ed25519:/tmp/id_ed25519  # Mount the SSH private key file
    environment:
      - DB_PORT=5432
      - DB_HOST=postgres
      - DB_NAME=database
      - DB_USERNAME=username
      - DB_PASSWORD=password
      ## SSH Configuration
      - SSH_HOST_NAME=hostname
      - SSH_PORT=22
      - SSH_USER=user
      - SSH_REMOTE_PATH=/home/jkaninda/backups
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

## Key Notes

- **Supported File Formats**: The restore process supports `.sql`, `.sql.gz`, `.sql.gpg`, and `.sql.gz.gpg` files.
- **SSH Path**: Use the `--path` flag to specify the folder on the SSH remote server where the backup file is located.
- **Encrypted Backups**: If the backup is encrypted with GPG, ensure the `GPG_PASSPHRASE` environment variable is set for automatic decryption.
- **SSH Authentication**: Use a private key (`SSH_IDENTIFY_FILE`) for SSH authentication instead of a password for better security.
- **Network Configuration**: Ensure the `pg-bkup` container is connected to the same network as your database.