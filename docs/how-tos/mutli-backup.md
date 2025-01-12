---
title: Run multiple database backup schedules in the same container
layout: default
parent: How Tos
nav_order: 11
---


# Multiple Backup Schedules

You can configure multiple backup schedules with different configurations by using a configuration file. 

This file can be mounted into the container at `/config/config.yaml`, `/config/config.yml`, or specified via the `BACKUP_CONFIG_FILE` environment variable.

---

## Configuration File

The configuration file allows you to define multiple databases and their respective backup settings. 

Below is an example configuration file:

```yaml
# Optional: Define a global cron expression for scheduled backups
# cronExpression: "@every 20m"
cronExpression: ""

databases:
  - host: postgres1
    port: 5432
    name: database1
    user: database1
    password: password
    path: /s3-path/database1  # For SSH or FTP, define the full path (e.g., /home/toto/backup/)
  
  - host: postgres2
    port: 5432
    name: lldap
    user: lldap
    password: password
    path: /s3-path/lldap  # For SSH or FTP, define the full path (e.g., /home/toto/backup/)
  
  - host: postgres3
    port: 5432
    name: keycloak
    user: keycloak
    password: password
    path: /s3-path/keycloak  # For SSH or FTP, define the full path (e.g., /home/toto/backup/)
  
  - host: postgres4
    port: 5432
    name: joplin
    user: joplin
    password: password
    path: /s3-path/joplin  # For SSH or FTP, define the full path (e.g., /home/toto/backup/)
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

## Key Notes

- **Global Cron Expression**: You can define a global `cronExpression` in the configuration file to schedule backups for all databases. If omitted, backups will run immediately.
- **Database-Specific Paths**: For SSH or FTP storage, ensure the `path` field contains the full remote path (e.g., `/home/toto/backup/`).
- **Environment Variables**: Use the `BACKUP_CONFIG_FILE` environment variable to specify the path to the configuration file.
- **Security**: Avoid hardcoding sensitive information like passwords in the configuration file. Use environment variables or secrets management tools instead.
