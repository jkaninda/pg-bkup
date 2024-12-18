---
title: Quickstart
layout: home
nav_order: 2
---

# Quickstart

## Simple Backup Using Docker CLI

To run a one-time backup, bind your local volume to `/backup` in the container and run the `backup` command:

```shell
docker run --rm --network your_network_name \
  -v $PWD/backup:/backup/ \
  -e "DB_HOST=dbhost" \
  -e "DB_USERNAME=username" \
  -e "DB_PASSWORD=password" \
  jkaninda/pg-bkup backup -d database_name
```

### Using a Full Configuration File

Alternatively, you can use an `--env-file` to pass a full configuration:

```shell
docker run --rm --network your_network_name \
  --env-file your-env-file \
  -v $PWD/backup:/backup/ \
  jkaninda/pg-bkup backup -d database_name
```

## Simple Backup Using Docker Compose

Here is an example `docker-compose.yml` configuration:

```yaml
services:
  pg-bkup:
    # In production, lock the image tag to a release version.
    # See https://github.com/jkaninda/pg-bkup/releases for available releases.
    image: jkaninda/pg-bkup
    container_name: pg-bkup
    command: backup
    volumes:
      - ./backup:/backup
    environment:
      - DB_PORT=5432
      - DB_HOST=postgres
      - DB_NAME=foo
      - DB_USERNAME=bar
      - DB_PASSWORD=password
      - TZ=Europe/Paris
    # Connect pg-bkup to the same network as your database.
    networks:
      - web

networks:
  web:
```

## Recurring Backup with Docker

To schedule recurring backups, use the `--cron-expression` flag:

```shell
docker run --rm --network network_name \
  -v $PWD/backup:/backup/ \
  -e "DB_HOST=hostname" \
  -e "DB_USERNAME=user" \
  -e "DB_PASSWORD=password" \
  jkaninda/pg-bkup backup -d dbName --cron-expression "@every 15m"
```

For predefined schedules, see the [documentation](https://jkaninda.github.io/pg-bkup/reference/#predefined-schedules).

## Backup Using Kubernetes

Here is an example Kubernetes `Job` configuration for backups:

```yaml
apiVersion: batch/v1
kind: Job
metadata:
  name: backup-job
spec:
  ttlSecondsAfterFinished: 100
  template:
    spec:
      containers:
        - name: pg-bkup
          # In production, lock the image tag to a release version.
          # See https://github.com/jkaninda/pg-bkup/releases for available releases.
          image: jkaninda/pg-bkup
          command:
            - /bin/sh
            - -c
            - backup -d dbname
          resources:
            limits:
              memory: "128Mi"
              cpu: "500m"
          env:
            - name: DB_HOST
              value: "postgres"
            - name: DB_USERNAME
              value: "postgres"
            - name: DB_PASSWORD
              value: "password"
          volumeMounts:
            - mountPath: /backup
              name: backup
      volumes:
        - name: backup
          hostPath:
            path: /home/toto/backup # Directory location on host
            type: Directory # Optional field
      restartPolicy: Never
```


