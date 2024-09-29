---
title: Restore database from AWS S3
layout: default
parent: How Tos
nav_order: 6
---

# Restore database from S3 storage

To restore the database, you need to add `restore` command and specify the file to restore by adding `--file store_20231219_022941.sql.gz`.

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
    command: restore --storage s3 -d my-database -f store_20231219_022941.sql.gz --path /my-custom-path
    volumes:
      - ./backup:/backup
    environment:
      - DB_PORT=5432
      - DB_HOST=postgres
      - DB_NAME=database
      - DB_USERNAME=username
      - DB_PASSWORD=password
      ## AWS configurations
      - AWS_S3_ENDPOINT=https://s3.amazonaws.com
      - AWS_S3_BUCKET_NAME=backup
      - AWS_REGION="us-west-2"
      - AWS_ACCESS_KEY=xxxx
      - AWS_SECRET_KEY=xxxxx
      ## In case you are using S3 alternative such as Minio and your Minio instance is not secured, you change it to true
      - AWS_DISABLE_SSL="false"
    # pg-bkup container must be connected to the same network with your database
    networks:
      - web
networks:
  web:
```
## Restore on Kubernetes

Simple Kubernetes restore Job:

```yaml
apiVersion: batch/v1
kind: Job
metadata:
  name: restore-db
spec:
  template:
    spec:
      containers:
        - name: pg-bkup
          image: jkaninda/pg-bkup
          command:
            - /bin/sh
            - -c
            - restore -s s3 --path /custom_path -f store_20231219_022941.sql.gz
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
  backoffLimit: 4
```
