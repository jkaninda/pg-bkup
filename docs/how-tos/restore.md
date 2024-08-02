---
title: Restore database
layout: default
parent: How Tos
nav_order: 4
---

# Restore database

To restore the database, you need to add `restore` subcommand to `pg-bkup` or `bkup` and specify the file to restore by adding `--file store_20231219_022941.sql.gz`.

{: .note }
It supports __.sql__ and __.sql.gz__ compressed file.

### Restore

```yml
services:
  pg-bkup:
    # In production, it is advised to lock your image tag to a proper
    # release version instead of using `latest`.
    # Check https://github.com/jkaninda/pg-bkup/releases
    # for a list of available releases.
    image: jkaninda/pg-bkup
    container_name: pg-bkup
    command:
      - /bin/sh
      - -c
      - pg-bkup restore -d database -f store_20231219_022941.sql.gz
    volumes:
      - ./backup:/backup
    environment:
      - DB_PORT=5432
      - DB_HOST=postgres
      - DB_NAME=database
      - DB_USERNAME=username
      - DB_PASSWORD=password
    # pg-bkup container must be connected to the same network with your database
    networks:
      - web
networks:
  web:
```