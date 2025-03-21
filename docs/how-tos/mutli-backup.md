---
title: Run multiple database backup schedules in the same container
layout: default
parent: How Tos
nav_order: 11
---


# Multiple Backup Schedules

This tool supports running multiple database backup schedules within the same container. 
You can configure these schedules with different settings using a **configuration file**. This flexibility allows you to manage backups for multiple databases efficiently.

---

## Configuration File Setup

The configuration file can be mounted into the container at `/config/config.yaml`, `/config/config.yml`, or specified via the `BACKUP_CONFIG_FILE` environment variable.

### Key Features:
- **Global Environment Variables**: Use these for databases that share the same configuration.
- **Database-Specific Overrides**: Override global settings for individual databases by specifying them in the configuration file or using the database name as a suffix in the variable name (e.g., `DB_HOST_DATABASE1`).
- **Global Cron Expression**: Define a global `cronExpression` in the configuration file to schedule backups for all databases. If omitted, backups will run immediately.
- **Configuration File Path**: Specify the configuration file path using:
    - The `BACKUP_CONFIG_FILE` environment variable.
    - The `--config` or `-c` flag for the backup command.

{: .note }
The bulk backup or migration process requires administrative privileges on the database.

---

## Configuration File Example

Below is an example configuration file (`config.yaml`) that defines multiple databases and their respective backup settings:

```yaml
# Optional: Define a global cron expression for scheduled backups.
# Example: "@every 20m" (runs every 20 minutes). If omitted, backups run immediately.
cronExpression: ""

databases:
  - host: postgres1       # Optional: Overrides DB_HOST or uses DB_HOST_DATABASE1.
    port: 5432            # Optional: Default is 5432. Overrides DB_PORT or uses DB_PORT_DATABASE1.
    name: database1       # Required: Database name.
    user: database1       # Optional: Overrides DB_USERNAME or uses DB_USERNAME_DATABASE1.
    password: password    # Optional: Overrides DB_PASSWORD or uses DB_PASSWORD_DATABASE1.
    path: /s3-path/database1  # Required: Backup path for SSH, FTP, or S3 (e.g., /home/toto/backup/).

  - host: postgres2       # Optional: Overrides DB_HOST or uses DB_HOST_LLAP.
    port: 5432            # Optional: Default is 5432. Overrides DB_PORT or uses DB_PORT_LLAP.
    name: lldap           # Required: Database name.
    user: lldap           # Optional: Overrides DB_USERNAME or uses DB_USERNAME_LLAP.
    password: password    # Optional: Overrides DB_PASSWORD or uses DB_PASSWORD_LLAP.
    path: /s3-path/lldap  # Required: Backup path for SSH, FTP, or S3 (e.g., /home/toto/backup/).

  - host: postgres3       # Optional: Overrides DB_HOST or uses DB_HOST_KEYCLOAK.
    port: 5432            # Optional: Default is 5432. Overrides DB_PORT or uses DB_PORT_KEYCLOAK.
    name: keycloak        # Required: Database name.
    user: keycloak        # Optional: Overrides DB_USERNAME or uses DB_USERNAME_KEYCLOAK.
    password: password    # Optional: Overrides DB_PASSWORD or uses DB_PASSWORD_KEYCLOAK.
    path: /s3-path/keycloak  # Required: Backup path for SSH, FTP, or S3 (e.g., /home/toto/backup/).

  - host: postgres4       # Optional: Overrides DB_HOST or uses DB_HOST_JOPLIN.
    port: 5432            # Optional: Default is 5432. Overrides DB_PORT or uses DB_PORT_JOPLIN.
    name: joplin          # Required: Database name.
    user: joplin          # Optional: Overrides DB_USERNAME or uses DB_USERNAME_JOPLIN.
    password: password    # Optional: Overrides DB_PASSWORD or uses DB_PASSWORD_JOPLIN.
    path: /s3-path/joplin  # Required: Backup path for SSH, FTP, or S3 (e.g., /home/toto/backup/).
```

---

## Docker Compose Configuration

To use the configuration file in a Docker Compose setup, mount the file and specify its path using the `BACKUP_CONFIG_FILE` environment variable.

### Example: Docker Compose File

```yaml
services:
  pg-bkup:
    # In production, lock your image tag to a specific release version
    # instead of using `latest`. Check https://github.com/jkaninda/pg-bkup/releases
    # for available releases.
    image: jkaninda/pg-bkup
    container_name: pg-bkup
    command: backup
    volumes:
      - ./backup:/backup  # Mount the backup directory
      - ./config.yaml:/backup/config.yaml  # Mount the configuration file
    environment:
      ## Specify the path to the configuration file
      - BACKUP_CONFIG_FILE=/backup/config.yaml
    # Ensure the pg-bkup container is connected to the same network as your database
    networks:
      - web

networks:
  web:
```

---



