---
title: Azure Blob storage
layout: default
parent: How Tos
nav_order: 5
---

# Backup to Azure Blob Storage

To store your backups on Azure Blob Storage, you can configure the backup process to use the `--storage azure` option. 

This section explains how to set up and configure Azure Blob-based backups.

---

## Configuration Steps

1. **Specify the Storage Type**  
   Add the `--storage azure` flag to your backup command.

2. **Set the Blob Path**  
   Optionally, specify a custom folder within your Azure Blob container where backups will be stored using the `--path` flag.  
   Example: `--path my-custom-path`.

3. **Required Environment Variables**  
   The following environment variables are mandatory for Azure Blob-based backups:

    - `AZURE_STORAGE_CONTAINER_NAME`: The name of the Azure Blob container where backups will be stored.
    - `AZURE_STORAGE_ACCOUNT_NAME`: The name of your Azure Storage account.
    - `AZURE_STORAGE_ACCOUNT_KEY`: The access key for your Azure Storage account.

---

## Example Configuration

Below is an example `docker-compose.yml` configuration for backing up to Azure Blob Storage:

```yaml
services:
  mysql-bkup:
    # In production, lock your image tag to a specific release version
    # instead of using `latest`. Check https://github.com/jkaninda/pg-bkup/releases
    # for available releases.
    image: jkaninda/pg-bkup
    container_name: pg-bkup
    command: backup --storage azure -d database --path my-custom-path
    environment:
      - DB_PORT=5432
      - DB_HOST=postgres
      - DB_NAME=database
      - DB_USERNAME=username
      - DB_PASSWORD=password
      ## Azure Blob Configuration
      - AZURE_STORAGE_CONTAINER_NAME=backup-container
      - AZURE_STORAGE_ACCOUNT_NAME=account-name
      - AZURE_STORAGE_ACCOUNT_KEY=Ppby8vdM02xNOcqFlqUwJPLlmEtlCDXJ1OUzFT50uSRZ6IFsuFq2UVErCz4I6tq/K1SZFPTOtr/KBHBeksoGMGw==

    # Ensure the pg-bkup container is connected to the same network as your database
    networks:
      - web

networks:
  web:
```

---

## Key Notes

- **Custom Path**: Use the `--path` flag to specify a folder within your Azure Blob container for organizing backups.
- **Security**: Ensure your `AZURE_STORAGE_ACCOUNT_KEY` is kept secure and not exposed in public repositories.
- **Compatibility**: This configuration works with Azure Blob Storage and other compatible storage solutions.
