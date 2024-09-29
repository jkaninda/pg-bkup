---
title: Migrate database
layout: default
parent: How Tos
nav_order: 10
---

# Migrate database

To migrate the database, you need to add `migrate` command.

{: .note }
The PostgresQL backup has another great feature: migrating your database from a source database to a target.

As you know, to restore a database from a source to a target database, you need 2 operations: which is to start by backing up the source database and then restoring the source backed database to the target database.
Instead of proceeding like that, you can use the integrated feature `(migrate)`, which will help you migrate your database by doing only one operation.

{: .warning }
The `migrate` operation is irreversible, please backup your target database before this action.

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
      ## Source database
      - DB_PORT=5432
      - DB_HOST=postgres
      - DB_NAME=database
      - DB_USERNAME=username
      - DB_PASSWORD=password
      ## Target database
      - TARGET_DB_HOST=target-postgres
      - TARGET_DB_PORT=5432
      - TARGET_DB_NAME=dbname
      - TARGET_DB_USERNAME=username
      - TARGET_DB_PASSWORD=password
    # mysql-bkup container must be connected to the same network with your database
    networks:
      - web
networks:
  web:
```


### Migrate database using Docker CLI


```
## Source database
DB_HOST=postgres
DB_PORT=5432
DB_NAME=dbname
DB_USERNAME=username
DB_PASSWORD=password

## Taget database
TARGET_DB_HOST=target-postgres
TARGET_DB_PORT=5432
TARGET_DB_NAME=dbname
TARGET_DB_USERNAME=username
TARGET_DB_PASSWORD=password
```

```shell
 docker run --rm --network your_network_name \
 --env-file your-env
 -v $PWD/backup:/backup/ \
 jkaninda/pg-bkup migrate
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
            - migrate
          resources:
            limits:
              memory: "128Mi"
              cpu: "500m"
          env:
            ## Source Database
            - name: DB_HOST
              value: "postgres"
            - name: DB_PORT
              value: "5432"
            - name: DB_NAME
              value: "dbname"
            - name: DB_USERNAME
              value: "username"
            - name: DB_PASSWORD
              value: "password"
            ## Target Database
            - name: TARGET_DB_HOST
              value: "target-postgres"
            - name: TARGET_DB_PORT
              value: "5432"
            - name: TARGET_DB_NAME
              value: "dbname"
            - name: TARGET_DB_USERNAME
              value: "username"
            - name: TARGET_DB_PASSWORD
              value: "password"
      restartPolicy: Never
```