---
title: Backup
layout: default
parent: How Tos
nav_order: 1
---

# Backup database

To backup the database, you need to add `backup` subcommand to `pg-bkup` or `bkup`.

{: .note }
The default storage is local storage mounted to __/backup__. The backup is compressed by default using gzip. The flag __`disable-compression`__ can be used when you need to disable backup compression.

{: .warning }
Creating a user for backup tasks who has read-only access is recommended!

The backup process can be run in scheduled mode for the recurring backups.
It handles __recurring__ backups of postgres database on Docker and can be deployed as __CronJob on Kubernetes__ using local, AWS S3 or SSH compatible storage.

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
      - pg-bkup backup -d database
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

### Backup using Docker CLI

```shell
 docker run --rm --network your_network_name \
 -v $PWD/backup:/backup/ \
 -e "DB_HOST=dbhost" \
 -e "DB_USERNAME=username" \
 -e "DB_PASSWORD=password" \
 jkaninda/pg-bkup  pg-bkup backup -d database_name
```

In case you need to use recurring backups, you can use `--mode scheduled` and specify the periodical backup time by adding `--period "0 1 * * *"` flag as described below.

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
      - pg-bkup backup -d database --mode scheduled --period "0 1 * * *"
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

