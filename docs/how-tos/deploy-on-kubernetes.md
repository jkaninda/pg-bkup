---
title: Deploy on Kubernetes
layout: default
parent: How Tos
nav_order: 9
---

# Deploy on Kubernetes

To deploy PostgreSQL Backup on Kubernetes, you can use a `Job` for one-time backups or restores, and a `CronJob` for recurring backups. Below are examples for different use cases.

---

## Backup Job to S3 Storage

This example demonstrates how to configure a Kubernetes `Job` to back up a PostgreSQL database to an S3-compatible storage.

```yaml
apiVersion: batch/v1
kind: Job
metadata:
  name: backup
spec:
  template:
    spec:
      containers:
      - name: pg-bkup
        # In production, lock your image tag to a specific release version
        # instead of using `latest`. Check https://github.com/jkaninda/pg-bkup/releases
        # for available releases.
        image: jkaninda/pg-bkup
        command: ["backup", "--storage", "s3"]
        resources:
          limits:
            memory: "128Mi"
            cpu: "500m"
        env:
          - name: DB_PORT
            value: "5432"
          - name: DB_HOST
            value: ""
          - name: DB_NAME
            value: ""
          - name: DB_USERNAME
            value: ""
          # Use Kubernetes Secrets for sensitive data like passwords
          - name: DB_PASSWORD
            value: ""
          - name: AWS_S3_ENDPOINT
            value: "https://s3.amazonaws.com"
          - name: AWS_S3_BUCKET_NAME
            value: "xxx"
          - name: AWS_REGION
            value: "us-west-2"
          - name: AWS_ACCESS_KEY
            value: "xxxx"
          - name: AWS_SECRET_KEY
            value: "xxxx"
          - name: AWS_DISABLE_SSL
            value: "false"
          - name: AWS_FORCE_PATH_STYLE
            value: "false"
      restartPolicy: Never
```

---

## Backup Job to SSH Remote Server

This example demonstrates how to configure a Kubernetes `Job` to back up a PostgreSQL database to an SSH remote server.

```yaml
apiVersion: batch/v1
kind: Job
metadata:
  name: backup
spec:
  ttlSecondsAfterFinished: 100
  template:
    spec:
      containers:
      - name: pg-bkup
        # In production, lock your image tag to a specific release version
        # instead of using `latest`. Check https://github.com/jkaninda/pg-bkup/releases
        # for available releases.
        image: jkaninda/pg-bkup
        command: ["backup", "--storage", "ssh", "--disable-compression"]
        resources:
          limits:
            memory: "128Mi"
            cpu: "500m"
        env:
          - name: DB_PORT
            value: "5432"
          - name: DB_HOST
            value: ""
          - name: DB_NAME
            value: "dbname"
          - name: DB_USERNAME
            value: "postgres"
          # Use Kubernetes Secrets for sensitive data like passwords
          - name: DB_PASSWORD
            value: ""
          - name: SSH_HOST_NAME
            value: "xxx"
          - name: SSH_PORT
            value: "22"
          - name: SSH_USER
            value: "xxx"
          - name: SSH_PASSWORD
            value: "xxxx"
          - name: SSH_REMOTE_PATH
            value: "/home/toto/backup"
          # Optional: Required if you want to encrypt your backup
          - name: GPG_PASSPHRASE
            value: "xxxx"
      restartPolicy: Never
```

---

## Restore Job

This example demonstrates how to configure a Kubernetes `Job` to restore a PostgreSQL database from a backup stored on an SSH remote server.

```yaml
apiVersion: batch/v1
kind: Job
metadata:
  name: restore-job
spec:
  ttlSecondsAfterFinished: 100
  template:
    spec:
      containers:
      - name: pg-bkup
        # In production, lock your image tag to a specific release version
        # instead of using `latest`. Check https://github.com/jkaninda/pg-bkup/releases
        # for available releases.
        image: jkaninda/pg-bkup
        command: ["restore", "--storage", "ssh", "--file", "store_20231219_022941.sql.gz"]
        resources:
          limits:
            memory: "128Mi"
            cpu: "500m"
        env:
          - name: DB_PORT
            value: "5432"
          - name: DB_HOST
            value: ""
          - name: DB_NAME
            value: "dbname"
          - name: DB_USERNAME
            value: "postgres"
          # Use Kubernetes Secrets for sensitive data like passwords
          - name: DB_PASSWORD
            value: ""
          - name: SSH_HOST_NAME
            value: "xxx"
          - name: SSH_PORT
            value: "22"
          - name: SSH_USER
            value: "xxx"
          - name: SSH_PASSWORD
            value: "xxxx"
          - name: SSH_REMOTE_PATH
            value: "/home/toto/backup"
          # Optional: Required if your backup was encrypted
          #- name: GPG_PASSPHRASE
          #  value: "xxxx"
      restartPolicy: Never
```

---

## Recurring Backup with CronJob

This example demonstrates how to configure a Kubernetes `CronJob` for recurring backups to an SSH remote server.

```yaml
apiVersion: batch/v1
kind: CronJob
metadata:
  name: backup-job
spec:
  schedule: "* * * * *"
  jobTemplate:
    spec:
      template:
        spec:
          containers:
          - name: pg-bkup
            # In production, lock your image tag to a specific release version
            # instead of using `latest`. Check https://github.com/jkaninda/pg-bkup/releases
            # for available releases.
            image: jkaninda/pg-bkup
            command: ["backup", "--storage", "ssh", "--disable-compression"]
            resources:
              limits:
                memory: "128Mi"
                cpu: "500m"
            env:
              - name: DB_PORT
                value: "5432"
              - name: DB_HOST
                value: ""
              - name: DB_NAME
                value: "test"
              - name: DB_USERNAME
                value: "postgres"
              # Use Kubernetes Secrets for sensitive data like passwords
              - name: DB_PASSWORD
                value: ""
              - name: SSH_HOST_NAME
                value: "192.168.1.16"
              - name: SSH_PORT
                value: "2222"
              - name: SSH_USER
                value: "jkaninda"
              - name: SSH_REMOTE_PATH
                value: "/config/backup"
              - name: SSH_PASSWORD
                value: "password"
              # Optional: Required if you want to encrypt your backup
              #- name: GPG_PASSPHRASE
              #  value: "xxx"
          restartPolicy: Never
```

---

## Kubernetes Rootless Deployment

This example demonstrates how to run the backup container in a rootless environment, suitable for platforms like OpenShift.

```yaml
apiVersion: batch/v1
kind: CronJob
metadata:
  name: backup-job
spec:
  schedule: "* * * * *"
  jobTemplate:
    spec:
      template:
        spec:
          securityContext:
            runAsUser: 1000
            runAsGroup: 3000
            fsGroup: 2000
          containers:
          - name: pg-bkup
            # In production, lock your image tag to a specific release version
            # instead of using `latest`. Check https://github.com/jkaninda/pg-bkup/releases
            # for available releases.
            image: jkaninda/pg-bkup
            command: ["backup", "--storage", "ssh", "--disable-compression"]
            resources:
              limits:
                memory: "128Mi"
                cpu: "500m"
            env:
              - name: DB_PORT
                value: "5432"
              - name: DB_HOST
                value: ""
              - name: DB_NAME
                value: "test"
              - name: DB_USERNAME
                value: "postgres"
              # Use Kubernetes Secrets for sensitive data like passwords
              - name: DB_PASSWORD
                value: ""
              - name: SSH_HOST_NAME
                value: "192.168.1.16"
              - name: SSH_PORT
                value: "2222"
              - name: SSH_USER
                value: "jkaninda"
              - name: SSH_REMOTE_PATH
                value: "/config/backup"
              - name: SSH_PASSWORD
                value: "password"
              # Optional: Required if you want to encrypt your backup
              #- name: GPG_PASSPHRASE
              #  value: "xxx"
          restartPolicy: OnFailure
```

---

## Migrate Database

This example demonstrates how to configure a Kubernetes `Job` to migrate a PostgreSQL database from one server to another.

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
        # In production, lock your image tag to a specific release version
        # instead of using `latest`. Check https://github.com/jkaninda/pg-bkup/releases
        # for available releases.
        image: jkaninda/pg-bkup
        command: ["migrate"]
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

---

## Key Notes

- **Security**: Always use Kubernetes Secrets for sensitive data like passwords and access keys.
- **Resource Limits**: Adjust resource limits (`memory` and `cpu`) based on your workload requirements.
- **Cron Schedule**: Use standard cron expressions for scheduling recurring backups.
- **Rootless Deployment**: The image supports running in rootless environments, making it suitable for platforms like OpenShift.
