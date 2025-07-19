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
# Optional: Global cron expression for scheduled backups.
# Examples: "@daily", "@every 5m", "0 3 * * *"
cronExpression: "@daily"

databases:
  - host: lldap-db             # Optional: Overrides DB_HOST or uses DB_HOST_LLDAP.
    port: 5432                 # Optional: Defaults to 5432. Overrides DB_PORT or uses DB_PORT_LLDAP.
    name: lldap                # Required: Database name
    user: lldap                # Optional: Can override via DB_USERNAME or uses DB_USERNAME_LLDAP.
    password: password         # Optional: Can override via DB_PASSWORD or uses DB_PASSWORD_LLDAP.
    path: /s3-path/lldap       # Required: Destination path (S3, FTP, SSH, etc.)

  - host: keycloak-db
    port: 5432
    name: keycloak
    user: keycloak
    password: password
    path: /s3-path/keycloak

  - host: gitea-db
    port: 5432
    name: gitea
    user: gitea
    password: ""               # Can be empty or sourced from DB_PASSWORD_GITEA
    path: /s3-path/gitea
```

> 🔹 **Tip:** You can override any field using environment variables. For example, `DB_PASSWORD_KEYCLOAK` takes precedence over the `password` field for the `keycloak` entry.

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
      - DB_PASSWORD_GITEA=password
    # Ensure the pg-bkup container is connected to the same network as your database
    networks:
      - web

networks:
  web:
```

---



