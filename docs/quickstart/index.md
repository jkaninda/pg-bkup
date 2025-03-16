---
title: Quickstart
layout: home
nav_order: 2
---

# Quickstart

This guide provides quick examples for running backups using Docker CLI, Docker Compose, and Kubernetes.

---

### Simple Backup Using Docker CLI

To perform a one-time backup, bind your local volume to `/backup` in the container and run the `backup` command:

```shell
docker run --rm --network your_network_name \
  -v $PWD/backup:/backup/ \
  -e "DB_HOST=dbhost" \
  -e "DB_PORT=5432" \
  -e "DB_USERNAME=username" \
  -e "DB_PASSWORD=password" \
  jkaninda/pg-bkup backup -d database_name
```

Alternatively, use an environment file (`--env-file`) for configuration:

```shell
docker run --rm --network your_network_name \
  --env-file your-env-file \
  -v $PWD/backup:/backup/ \
  jkaninda/pg-bkup backup -d database_name
```

### Backup All Databases

To back up all databases on the server, use the `--all-databases` or `-a` flag. By default, this creates individual backup files for each database.

```shell
docker run --rm --network your_network_name \
  -v $PWD/backup:/backup/ \
  -e "DB_HOST=dbhost" \
  -e "DB_PORT=5432" \
  -e "DB_USERNAME=username" \
  -e "DB_PASSWORD=password" \
  jkaninda/pg-bkup backup --all-databases --disable-compression
```

> **Note:** Use the `--all-in-one` or `-A` flag to combine backups into a single file.

---

### Simple Restore Using Docker CLI

To restore a database, bind your local volume to `/backup` and run the `restore` command:

```shell
docker run --rm --network your_network_name \
  -v $PWD/backup:/backup/ \
  -e "DB_HOST=dbhost" \
  -e "DB_PORT=5432" \
  -e "DB_USERNAME=username" \
  -e "DB_PASSWORD=password" \
  jkaninda/pg-bkup restore -d database_name -f backup_file.sql.gz
```

---

### Backup with Docker Compose

Below is an example of a `docker-compose.yml` file for running a one-time backup:

```yaml
services:
  pg-bkup:
    # In production, pin your image tag to a specific release version instead of `latest`.
    # See available releases: https://github.com/jkaninda/pg-bkup/releases
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
    networks:
      - web

networks:
  web:
```

---

### Recurring Backups with Docker

You can schedule recurring backups using the `--cron-expression` or `-e` flag:

```shell
docker run --rm --network network_name \
  -v $PWD/backup:/backup/ \
  -e "DB_HOST=hostname" \
  -e "DB_USERNAME=user" \
  -e "DB_PASSWORD=password" \
  jkaninda/pg-bkup backup -d dbName --cron-expression "@every 15m"
```

For predefined schedules, refer to the [documentation](https://jkaninda.github.io/pg-bkup/reference/#predefined-schedules).

---

## Deploy on Kubernetes

For Kubernetes, you can deploy `pg-bkup` as a Job or CronJob. Below are examples for both.

### Kubernetes Backup Job

This example defines a one-time backup job:

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
          # Pin the image tag to a specific release version in production.
          # See available releases: https://github.com/jkaninda/pg-bkup/releases
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
            path: /home/toto/backup # Directory location on the host
            type: Directory # Optional field
      restartPolicy: Never
```

### Kubernetes CronJob for Scheduled Backups

For scheduled backups, use a `CronJob`:

```yaml
apiVersion: batch/v1
kind: CronJob
metadata:
  name: pg-bkup-cronjob
spec:
  schedule: "0 2 * * *" # Runs daily at 2 AM
  jobTemplate:
    spec:
      template:
        spec:
          containers:
            - name: pg-bkup
              image: jkaninda/pg-bkup
              command:
                - /bin/sh
                - -c
                - backup -d dbname
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
                path: /home/toto/backup
                type: Directory
          restartPolicy: OnFailure
```
---

## Key Notes

- **Volume Binding**: Ensure the `/backup` directory is mounted to persist backup files.
- **Environment Variables**: Use environment variables or an `--env-file` to pass database credentials and other configurations.
- **Cron Expressions**: Use standard cron expressions or predefined schedules for recurring backups.
- **Kubernetes Jobs**: Use Kubernetes `Job` or `CronJob` for running backups in a Kubernetes cluster.