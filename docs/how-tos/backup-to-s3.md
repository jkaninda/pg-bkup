---
title: Backup to AWS S3
layout: default
parent: How Tos
nav_order: 2
---
# Backup to AWS S3 

{: .note }
As described on local backup section, to change the storage of you backup and use S3 as storage. You need to add `--storage s3` (-s s3).
You can also specify a specify folder where you want to save you data by adding `--path /my-custom-path` flag.


## Backup to S3

```yml
services:
  pg-bkup:
    # In production, it is advised to lock your image tag to a proper
    # release version instead of using `latest`.
    # Check https://github.com/jkaninda/pg-bkup/releases
    # for a list of available releases.
    image: jkaninda/pg-bkup
    container_name: pg-bkup
    command: backup --storage s3 -d database --path /my-custom-path
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
      - AWS_FORCE_PATH_STYLE="false"
 
    # pg-bkup container must be connected to the same network with your database
    networks:
      - web
networks:
  web:
```

### Recurring backups to S3

As explained above, you need just to add AWS environment variables and specify the storage type `--storage s3`.
In case you need to use recurring backups, you can use `--cron-expression "0 1 * * *"` flag or  `BACKUP_CRON_EXPRESSION=0 1 * * *` as described below.

```yml
services:
  pg-bkup:
    # In production, it is advised to lock your image tag to a proper
    # release version instead of using `latest`.
    # Check https://github.com/jkaninda/pg-bkup/releases
    # for a list of available releases.
    image: jkaninda/pg-bkup
    container_name: pg-bkup
    command: backup --storage s3 -d my-database --mode scheduled --period "0 1 * * *"
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
     # - BACKUP_CRON_EXPRESSION=0 1 * * * # Optional
      ## In case you are using S3 alternative such as Minio and your Minio instance is not secured, you change it to true
      - AWS_DISABLE_SSL="false"
     # pg-bkup container must be connected to the same network with your database
    networks:
      - web
networks:
  web:
```

