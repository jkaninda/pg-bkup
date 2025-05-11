# PG-BKUP

**PG-BKUP** is a Docker container image designed to **backup, restore, and migrate PostgreSQL databases**.
It supports a variety of storage options and ensures data security through GPG encryption.

PG-BKUP is designed for seamless deployment on **Docker** and **Kubernetes**, simplifying PostgreSQL backup, restoration, and migration across environments.
It is a lightweight, multi-architecture solution compatible with **Docker**, **Docker Swarm**, **Kubernetes**, and other container orchestration platforms.

[![Tests](https://github.com/jkaninda/pg-bkup/actions/workflows/tests.yml/badge.svg)](https://github.com/jkaninda/pg-bkup/actions/workflows/tests.yml)
[![Build](https://github.com/jkaninda/pg-bkup/actions/workflows/release.yml/badge.svg)](https://github.com/jkaninda/pg-bkup/actions/workflows/release.yml)
[![Go Report](https://goreportcard.com/badge/github.com/jkaninda/mysql-bkup)](https://goreportcard.com/report/github.com/jkaninda/pg-bkup)
![Docker Image Size (latest by date)](https://img.shields.io/docker/image-size/jkaninda/pg-bkup?style=flat-square)
![Docker Pulls](https://img.shields.io/docker/pulls/jkaninda/pg-bkup?style=flat-square)
<a href="https://ko-fi.com/jkaninda"><img src="https://uploads-ssl.webflow.com/5c14e387dab576fe667689cf/5cbed8a4ae2b88347c06c923_BuyMeACoffee_blue.png" height="20" alt="buy ma a coffee"></a>

## Features

- **Flexible Storage Backends:**
    - Local filesystem
    - Amazon S3 & S3-compatible storage (e.g., MinIO, Wasabi)
    - FTP
    - SSH-compatible storage
    - Azure Blob storage

- **Data Security:**
    - Backups can be encrypted using **GPG** to ensure confidentiality.

- **Deployment Flexibility:**
    - Available as the [jkaninda/pg-bkup](https://hub.docker.com/r/jkaninda/pg-bkup) Docker image.
    - Deployable on **Docker**, **Docker Swarm**, and **Kubernetes**.
    - Supports recurring backups of PostgreSQL databases when deployed:
        - On Docker for automated backup schedules.
        - As a **Job** or **CronJob** on Kubernetes.

- **Notifications:**
    - Get real-time updates on backup success or failure via:
        - **Telegram**
        - **Email**

## 💡Use Cases

- **Scheduled Backups**: Automate recurring backups using Docker or Kubernetes.
- **Disaster Recovery:** Quickly restore backups to a clean PostgreSQL instance.
- **Database Migration**: Seamlessly move data across environments using the built-in `migrate` feature.
- **Secure Archiving:** Keep backups encrypted and safely stored in the cloud or remote servers.


## ✅ Verified Platforms
PG-BKUP has been tested and runs successfully on:

- Docker
- Docker Swarm
- Kubernetes
- OpenShift

## Documentation is found at <https://jkaninda.github.io/pg-bkup>


## Useful links

- [Docker Hub](https://hub.docker.com/r/jkaninda/pg-bkup)
- [Github](https://github.com/jkaninda/pg-bkup)

### MySQL solution :

- [MySQL](https://github.com/jkaninda/mysql-bkup)


## Quickstart

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

### Migrate database

The `migrate` command allows you to transfer a PostgreSQL database from a source to a target database in a single step, combining backup and restore operations.


```bash
docker run --rm --network your_network_name \
  --env-file your-env \
  jkaninda/pg-bkup migrate
```

>  **Note:** Use the `--all-databases` (`-a`) flag to migrate all databases, PG-BKUP supports database creation if it does not exist on the target database.


For database migration, refer to the [documentation](https://jkaninda.github.io/pg-bkup/how-tos/migrate.html).


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
          command: ["backup", "-d", "db_name"]
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
              command: ["backup", "-d", "db_name"]
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

## 🚀 Why Use PG-BKUP?

**PG-BKUP** isn't just another PostgreSQL backup tool, it's a robust, production-ready solution purpose-built for modern DevOps workflows.

Here’s why developers, sysadmins, and DevOps choose **PG-BKUP**:

### ✅ All-in-One Backup, Restore & Migration

Whether you're backing up a single database, restoring critical data, or migrating across environments, PG-BKUP handles it all with a **single, unified CLI** no scripting gymnastics required.


### 🔄 Works Everywhere You Deploy

Designed to be cloud-native:

* **Runs seamlessly on Docker, Docker Swarm, and Kubernetes**
* Supports **CronJobs** for automated scheduled backups
* Compatible with GitOps and CI/CD workflows

### ☁️ Flexible Storage Integrations

Store your backups **anywhere**:

* Local disks
* Amazon S3, MinIO, Wasabi, Azure Blob, FTP, SSH

### 🔒 Enterprise-Grade Security

* **GPG Encryption**: Protect sensitive data with optional encryption before storing backups locally or in the cloud.
* **Secure Storage** Options: Supports S3, Azure Blob, SFTP, and SSH with encrypted transfers, keeping backups safe from unauthorized access.

### 📬 Instant Notifications

Stay in the loop with real-time notifications via **Telegram** and **Email**. Know immediately when a backup succeeds—or fails.

### 🏃‍♂️ Lightweight and Fast

Written in **Go**, PG-BKUP is fast, multi-arch compatible (`amd64`, `arm64`, `arm/v7`), and optimized for minimal memory and CPU usage. Ideal for both cloud and edge deployments.

### 🧪 Tested. Verified. Trusted.

Actively maintained with **automated testing**, **Docker image size optimizations**, and verified support across major container platforms.

---

## Available image registries

This Docker image is published to both Docker Hub and the GitHub container registry.
Depending on your preferences and needs, you can reference both `jkaninda/pg-bkup` as well as `ghcr.io/jkaninda/pg-bkup`:

```
docker pull jkaninda/pg-bkup
docker pull ghcr.io/jkaninda/pg-bkup
```

Documentation references Docker Hub, but all examples will work using ghcr.io just as well.

## References

We created this image as a simpler and more lightweight alternative to existing solutions. Here’s why:

- **Lightweight:** Written in Go, the image is optimized for performance and minimal resource usage.
- **Multi-Architecture Support:** Supports `arm64` and `arm/v7` architectures.
- **Docker Swarm Support:** Fully compatible with Docker in Swarm mode.
- **Kubernetes Support:** Designed to work seamlessly with Kubernetes.



## License

This project is licensed under the MIT License. See the LICENSE file for details.

## Authors

**Jonas Kaninda**
- <https://github.com/jkaninda>

## Copyright

Copyright (c) [2023] [Jonas Kaninda]
