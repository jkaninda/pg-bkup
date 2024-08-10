---
title: Deploy on Kubernetes
layout: default
parent: How Tos
nav_order: 8
---

## Deploy on Kubernetes

To deploy PostgreSQL Backup on Kubernetes, you can use Job to backup or Restore your database.
For recurring backup you can use CronJob, you don't need to run it in scheduled mode. as described bellow.

## Backup Job

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
        - bkup
        - backup
        - --storage
        - ssh
        - --disable-compression
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
        - bkup
        - restore
        - --storage
        - ssh
        - --file store_20231219_022941.sql.gz
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
            - bkup
            - backup
            - --storage
            - ssh
            - --disable-compression
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
                - bkup
                - backup
                - --storage
                - ssh
                - --disable-compression
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
