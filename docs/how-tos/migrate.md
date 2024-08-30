---
title: Migrate database
layout: default
parent: How Tos
nav_order: 9
---

# Migrate database

To migrate the database, you need to add `migrate` command.

{: .note }
The pg backup has another great feature: migrating your database from a source database to another.

As you know, to restore a database from a source to a target database, you need 2 operations: which is to start by backing up the source database and then restoring the source backed database to the target database.
Instead of proceeding like that, you can use the integrated feature `(migrate)`, which will help you migrate your database by doing only one operation.


### Docker compose
```yml
services:
  pg-bkup:
    # In production, it is advised to lock your image tag to a proper
    # release version instead of using `latest`.
    # Check https://github.com/jkaninda/pg-bkup/releases
    # for a list of available releases.
    image: jkaninda/pg-bkup
    container_name: pg-bkup
    command: migrate
    volumes:
      - ./backup:/backup
    environment:
      ## Target database
      - DB_PORT=5432
      - DB_HOST=postgres
      - DB_NAME=database
      - DB_USERNAME=username
      - DB_PASSWORD=password
      ## Source database
      - SOURCE_DB_HOST=postgres
      - SOURCE_DB_PORT=5432
      - SOURCE_DB_NAME=sourcedb
      - SOURCE_DB_USERNAME=jonas
      - SOURCE_DB_PASSWORD=password
    # pg-bkup container must be connected to the same network with your database
    networks:
      - web
networks:
  web:
```

### Migrate database using Docker CLI

```env
## Target database
DB_PORT=5432
DB_HOST=postgres
DB_NAME=targetdb
DB_USERNAME=targetuser
DB_PASSWORD=password

## Source database
SOURCE_DB_HOST=postgres
SOURCE_DB_PORT=5432
SOURCE_DB_NAME=sourcedb
SOURCE_DB_USERNAME=sourceuser
SOURCE_DB_PASSWORD=password
```

```shell
 docker run --rm --network your_network_name \
 --env-file your-env
 -v $PWD/backup:/backup/ \
 jkaninda/pg-bkup migrate -d database_name
```

## Kubernetes

```yaml
apiVersion: batch/v1
kind: Job
metadata:
  name: migrate-db
spec:
  ttlSecondsAfterFinished: 100
  template:
    spec:
      containers:
      - name: pg-bkup
        # In production, it is advised to lock your image tag to a proper
        # release version instead of using `latest`.
        # Check https://github.com/jkaninda/pg-bkup/releases
        # for a list of available releases.
        image: jkaninda/pg-bkup
        command:
        - /bin/sh
        - -c
        - migrate -d targetdb
        resources:
          limits:
            memory: "128Mi"
            cpu: "500m"
        env:
        ## Target DB
          - name: DB_HOST
            value: "postgres-target"
          - name: DB_USERNAME
            value: "postgres"
          - name: DB_PASSWORD
            value: "password"
          ## Source DB
          - name: SOURCE_DB_HOST
            value: "postgres-source"
          - name: SOURCE_DB_NAME
            value: "sourcedb"
          - name: SOURCE_DB_USERNAME
            value: "postgres"
          # Please use secret!
          - name: SOURCE_DB_PASSWORD
            value: "password"
      restartPolicy: Never
```