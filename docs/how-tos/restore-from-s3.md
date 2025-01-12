---
title: Restore database from AWS S3
layout: default
parent: How Tos
nav_order: 6
---

# Restore Database from S3 Storage

To restore a PostgreSQL database from a backup stored in S3, use the `restore` command and specify the backup file with the `--file` flag. The system supports the following file formats:

- `.sql` (uncompressed SQL dump)
- `.sql.gz` (gzip-compressed SQL dump)
- `.sql.gpg` (GPG-encrypted SQL dump)
- `.sql.gz.gpg` (GPG-encrypted and gzip-compressed SQL dump)

---

## Configuration Steps

1. **Specify the Backup File**: Use the `--file` flag to specify the backup file to restore.
2. **Set the Storage Type**: Add the `--storage s3` flag to indicate that the backup is stored in S3.
3. **Provide S3 Configuration**: Include the necessary AWS S3 credentials and configuration.
4. **Provide Database Credentials**: Ensure the correct database connection details are provided.

---

## Example: Restore from S3 Configuration

Below is an example `docker-compose.yml` configuration for restoring a database from S3 storage:

```yaml
services:
  pg-bkup:
    # In production, lock your image tag to a specific release version
    # instead of using `latest`. Check https://github.com/jkaninda/pg-bkup/releases
    # for available releases.
    image: jkaninda/pg-bkup
    container_name: pg-bkup
    command: restore --storage s3 -d my-database -f store_20231219_022941.sql.gz --path /my-custom-path
    volumes:
      - ./backup:/backup  # Mount the directory for local operations (if needed)
    environment:
      - DB_PORT=5432
      - DB_HOST=postgres
      - DB_NAME=database
      - DB_USERNAME=username
      - DB_PASSWORD=password
      ## AWS S3 Configuration
      - AWS_S3_ENDPOINT=https://s3.amazonaws.com
      - AWS_S3_BUCKET_NAME=backup
      - AWS_REGION=us-west-2
      - AWS_ACCESS_KEY=xxxx
      - AWS_SECRET_KEY=xxxxx
      ## Optional: Disable SSL for S3 alternatives like Minio
      - AWS_DISABLE_SSL=false
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

- **Supported File Formats**: The restore process supports `.sql`, `.sql.gz`, `.sql.gpg`, and `.sql.gz.gpg` files.
- **S3 Path**: Use the `--path` flag to specify the folder within the S3 bucket where the backup file is located.
- **Encrypted Backups**: If the backup is encrypted with GPG, ensure the `GPG_PASSPHRASE` environment variable is set for automatic decryption.
- **S3 Alternatives**: For S3-compatible storage like Minio, set `AWS_DISABLE_SSL` and `AWS_FORCE_PATH_STYLE` as needed.
- **Network Configuration**: Ensure the `pg-bkup` container is connected to the same network as your database.