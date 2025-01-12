---
title: Restore database
layout: default
parent: How Tos
nav_order: 5
---


# Restore Database

To restore a PostgreSQL database, use the `restore` command and specify the backup file to restore with the `--file` flag. The system supports the following file formats:

- `.sql` (uncompressed SQL dump)
- `.sql.gz` (gzip-compressed SQL dump)
- `.sql.gpg` (GPG-encrypted SQL dump)
- `.sql.gz.gpg` (GPG-encrypted and gzip-compressed SQL dump)

---

## Configuration Steps

1. **Specify the Backup File**: Use the `--file` flag to specify the backup file to restore.
2. **Provide Database Credentials**: Ensure the correct database connection details are provided.

---

## Example: Restore Configuration

Below is an example `docker-compose.yml` configuration for restoring a database:

```yaml
services:
  pg-bkup:
    # In production, lock your image tag to a specific release version
    # instead of using `latest`. Check https://github.com/jkaninda/pg-bkup/releases
    # for available releases.
    image: jkaninda/pg-bkup
    container_name: pg-bkup
    command: restore -d database -f store_20231219_022941.sql.gz
    volumes:
      - ./backup:/backup  # Mount the directory containing the backup file
    environment:
      - DB_PORT=5432
      - DB_HOST=postgres
      - DB_NAME=database
      - DB_USERNAME=username
      - DB_PASSWORD=password
    # Ensure the pg-bkup container is connected to the same network as your database
    networks:
      - web

networks:
  web:
```

---

## Key Notes

- **Supported File Formats**: The restore process supports `.sql`, `.sql.gz`, `.sql.gpg`, and `.sql.gz.gpg` files.
- **Encrypted Backups**: If the backup is encrypted with GPG, ensure the `GPG_PASSPHRASE` environment variable is set for automatic decryption.
- **Network Configuration**: Ensure the `pg-bkup` container is connected to the same network as your database.
