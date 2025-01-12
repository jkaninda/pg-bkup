---
title: Backup to FTP remote server
layout: default
parent: How Tos
nav_order: 4
---

# Backup to FTP Remote Server

To store your backups on an FTP remote server, you can configure the backup process to use the `--storage ftp` option. This section explains how to set up and configure FTP-based backups.

---

## Configuration Steps

1. **Specify the Storage Type**  
   Add the `--storage ftp` flag to your backup command.

2. **Set the Remote Path**  
   Define the full remote path where backups will be stored using the `--path` flag or the `REMOTE_PATH` environment variable.  
   Example: `--path /home/jkaninda/backups`.

3. **Required Environment Variables**  
   The following environment variables are mandatory for FTP-based backups:

    - `FTP_HOST`: The hostname or IP address of the FTP server.
    - `FTP_PORT`: The FTP port (default is `21`).
    - `FTP_USER`: The username for FTP authentication.
    - `FTP_PASSWORD`: The password for FTP authentication.
    - `REMOTE_PATH`: The directory on the FTP server where backups will be stored.

---

## Example Configuration

Below is an example `docker-compose.yml` configuration for backing up to an FTP remote server:

```yaml
services:
  pg-bkup:
    # In production, lock your image tag to a specific release version
    # instead of using `latest`. Check https://github.com/jkaninda/pg-bkup/releases
    # for available releases.
    image: jkaninda/pg-bkup
    container_name: pg-bkup
    command: backup --storage ftp -d database
    environment:
      - DB_PORT=5432
      - DB_HOST=postgres
      - DB_NAME=database
      - DB_USERNAME=username
      - DB_PASSWORD=password
      ## FTP Configuration
      - FTP_HOST="hostname"
      - FTP_PORT=21
      - FTP_USER=user
      - FTP_PASSWORD=password
      - REMOTE_PATH=/home/jkaninda/backups

    # Ensure the pg-bkup container is connected to the same network as your database
    networks:
      - web

networks:
  web:
```

---

## Key Notes

- **Security**: FTP transmits data, including passwords, in plaintext. For better security, consider using SFTP (SSH File Transfer Protocol) or FTPS (FTP Secure) if supported by your server.
- **Remote Path**: Ensure the `REMOTE_PATH` directory exists on the FTP server and is writable by the specified `FTP_USER`.