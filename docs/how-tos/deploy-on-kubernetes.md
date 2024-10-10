---
title: Deploy on Kubernetes
layout: default
parent: How Tos
nav_order: 9
---

## Deploy on Kubernetes

To deploy PostgreSQL Backup on Kubernetes, you can use Job to backup or Restore your database.
For recurring backup you can use CronJob, you don't need to run it in scheduled mode. as described bellow.

## Backup Job to S3 Storage

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
        # In production, it is advised to lock your image tag to a proper
        # release version instead of using `latest`.
        # Check https://github.com/jkaninda/pg-bkup/releases
        # for a list of available releases.
        image: jkaninda/pg-bkup
        command:
        - /bin/sh
        - -c
        - backup --storage s3
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
          # Please use secret!
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
      restartPolicy: Never
```

## Backup Job to SSH remote Server

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
        # In production, it is advised to lock your image tag to a proper
        # release version instead of using `latest`.
        # Check https://github.com/jkaninda/pg-bkup/releases
        # for a list of available releases.
        image: jkaninda/pg-bkup
        command:
        - /bin/sh
        - -c
        - backup --storage ssh --disable-compression
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
          # Please use secret!
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
          # Optional, required if you want to encrypt your backup
          - name: GPG_PASSPHRASE
            value: "xxxx"
      restartPolicy: Never
```

## Restore Job

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
        # In production, it is advised to lock your image tag to a proper
        # release version instead of using `latest`.
        # Check https://github.com/jkaninda/pg-bkup/releases
        # for a list of available releases.
        image: jkaninda/pg-bkup
        command:
        - /bin/sh
        - -c
        - restore --storage ssh --file store_20231219_022941.sql.gz
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
        # Please use secret!
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
          # Optional, required if your backup was encrypted
        #- name: GPG_PASSPHRASE
        #  value: "xxxx"
      restartPolicy: Never
```

## Recurring backup

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
            # In production, it is advised to lock your image tag to a proper
            # release version instead of using `latest`.
            # Check https://github.com/jkaninda/pg-bkup/releases
            # for a list of available releases.
            image: jkaninda/pg-bkup
            command:
            - /bin/sh
            - -c
            - backup --storage ssh --disable-compression
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
            # Please use secret!
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
            # Optional, required if you want to encrypt your backup
            #- name: GPG_PASSPHRASE
            #  value: "xxx"
          restartPolicy: Never
```

## Kubernetes Rootless

This image also supports Kubernetes security context, you can run it in Rootless environment.
It has been tested on Openshift, it works well.

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
              # In production, it is advised to lock your image tag to a proper
              # release version instead of using `latest`.
              # Check https://github.com/jkaninda/pg-bkup/releases
              # for a list of available releases.
              image: jkaninda/pg-bkup
              command:
                - /bin/sh
                - -c
                - backup --storage ssh --disable-compression
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
                # Please use secret!
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
              # Optional, required if you want to encrypt your backup
              #- name: GPG_PASSPHRASE
              #  value: "xxx"
          restartPolicy: OnFailure
```

## Migrate database

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
