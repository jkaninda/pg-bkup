---
title: Azure Blob storage
layout: default
parent: How Tos
nav_order: 5
---
# Azure Blob storage

{: .note }
As described on local backup section, to change the storage of your backup and use Azure Blob as storage. You need to add `--storage azure` (-s azure).
You can also specify a folder where you want to save you data by adding `--path my-custom-path` flag.


## Backup to Azure Blob storage

```yml
services:
  mysql-bkup:
    # In production, it is advised to lock your image tag to a proper
    # release version instead of using `latest`.
    # Check https://github.com/jkaninda/pg-bkup/releases
    # for a list of available releases.
    image: jkaninda/pg-bkup
    container_name: pg-bkup
    command: backup --storage azure -d database --path my-custom-path
    environment:
      - DB_PORT=5432
      - DB_HOST=postgres
      - DB_NAME=database
      - DB_USERNAME=username
      - DB_PASSWORD=password
      ## Azure Blob configurations
      - AZURE_STORAGE_CONTAINER_NAME=backup-container
      - AZURE_STORAGE_ACCOUNT_NAME=account-name
      - AZURE_STORAGE_ACCOUNT_KEY=Ppby8vdM02xNOcqFlqUwJPLlmEtlCDXJ1OUzFT50uSRZ6IFsuFq2UVErCz4I6tq/K1SZFPTOtr/KBHBeksoGMGw==
    # pg-bkup container must be connected to the same network with your database
    networks:
      - web
networks:
  web:
```



