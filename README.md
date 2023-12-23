# Postgres Backup
Postgres Backup tool, backup database to S3 or Object Storage

[![Build](https://github.com/jkaninda/pg-bkup/actions/workflows/build.yml/badge.svg)](https://github.com/jkaninda/pg-bkup/actions/workflows/build.yml)
![Docker Image Size (latest by date)](https://img.shields.io/docker/image-size/jkaninda/pg-bkup?style=flat-square)
![Docker Pulls](https://img.shields.io/docker/pulls/jkaninda/pg-bkup?style=flat-square)

- Docker
- Kubernetes

> Links:

- [Docker Hub](https://hub.docker.com/r/jkaninda/pg-bkup)
- [Github](https://github.com/jkaninda/pg-bkup)

> MySQL solution :
- [MySQL](https://github.com/jkaninda/mysql-bkup)

## Storage:
- local
- s3
- Object storage
## Usage

| Options       | Shorts | Usage                              |
|---------------|--------|------------------------------------|
| pg_bkup    | bkup   | CLI utility                    |
| --operation   | -o     | Set operation. backup or restore (default: backup)    |
| --storage      | -s     | Set storage. local or s3 (default: local)        |
| --file        | -f     | Set file name for restoration      |
| --path        |      | Set s3 path without file name. eg: /custom_path      |
| --dbname        | -d     | Set database name      |
| --port        | -p     | Set database port (default: 3306)      |
| --timeout     | -t     | Set timeout (default: 60s)        |
| --help        | -h     | Print this help message and exit   |
| --version     | -V     | Print version information and exit |

## Backup database :

Simple backup usage

```sh
bkup --operation backup
```
```sh
bkup -o backup
```
### S3

```sh
bkup --operation backup --storage s3
```
## Docker run:

```sh
docker run --rm --network your_network_name --name pg-bkup -v $PWD/backup:/backup/ -e "DB_HOST=database_host_name" -e "DB_USERNAME=username" -e "DB_PASSWORD=password" jkaninda/pg-bkup  bkup -o backup -d database_name
```

## Docker compose file:
```yaml
version: '3'
services:
  postgres:
    image: postgres:14.5
    container_name: postgres
    pull_policy: if_not_present
    restart: unless-stopped
    volumes:
      - ./postgres:/var/lib/postgresql/data
    environment:
      POSTGRES_DB: bkup
      POSTGRES_PASSWORD: password
      POSTGRES_USER: bkup
  pg-bkup:
    image: jkaninda/pg-bkup
    container_name: pg-bkup
    depends_on:
      - postgres
    command:
      - /bin/sh
      - -c
      - bkup --operation backup -d bkup
    volumes:
      - ./backup:/backup
    environment:
      - DB_PORT=5432
      - DB_HOST=postgres
      - DB_NAME=bkup
      - DB_USERNAME=bkup
      - DB_PASSWORD=password
```
## Restore database :

Simple database restore operation usage

```sh
bkup --operation restore --file database_20231217_115621.sql  --dbname database_name
```

```sh
bkup -o restore -f database_20231217_115621.sql -d database_name
```
### S3

```sh
bkup --operation restore --storage s3 --file database_20231217_115621.sql --dbname database_name
```

## Docker run:

```sh
docker run --rm --network your_network_name --name pg-bkup -v $PWD/backup:/backup/ -e "DB_HOST=database_host_name" -e "DB_USERNAME=username" -e "DB_PASSWORD=password" jkaninda/pg-bkup  bkup -o restore -d database_name -f napata_20231219_022941.sql.gz
```

## Docker compose file:

```yaml
version: '3'
services:
  pg-bkup:
    image: jkaninda/pg-bkup:latest
    container_name: pg-bkup
    command:
      - /bin/sh
      - -c
      - bkup --operation restore --file database_20231217_115621.sql -d database_name
    volumes:
      - ./backup:/backup
    environment:
      #- FILE_NAME=database_20231217_040238.sql.gz # Optional if file name is set from command
      - DB_PORT=5432
      - DB_HOST=postgres
      - DB_USERNAME=user_name
      - DB_PASSWORD=password
```
## Run 

```sh
docker-compose up -d
```
## Backup to S3

```sh
docker run --rm --privileged --device /dev/fuse --name pg-bkup -e "DB_HOST=db_hostname" -e "DB_USERNAME=username" -e "DB_PASSWORD=password" -e "ACCESS_KEY=your_access_key" -e "SECRET_KEY=your_secret_key" -e "BUCKETNAME=your_bucket_name" -e "S3_ENDPOINT=https://eu2.contabostorage.com" jkaninda/pg-bkup  bkup -o backup -s s3 -d database_name
```
> To change s3 backup path add this flag : --path /mycustomPath . default path is /pg_bkup

Simple S3 backup usage

```sh
bkup --operation backup --storage s3 --dbname mydatabase 
```
```yaml
  mysql-bkup:
    image: jkaninda/pg-bkup:latest
    container_name: pg-bkup
    tty: true
    privileged: true
    devices:
    - "/dev/fuse"
    command:
      - /bin/sh
      - -c
      - pg_bkup --operation restore --storage s3 -f database_20231217_115621.sql.gz --dbname database_name
    environment:
      - DB_PORT=5432
      - DB_HOST=postgress
      - DB_USERNAME=user_name
      - DB_PASSWORD=password
      - ACCESS_KEY=${ACCESS_KEY}
      - SECRET_KEY=${SECRET_KEY}
      - BUCKETNAME=${BUCKETNAME}
      - S3_ENDPOINT=${S3_ENDPOINT}

```
## Run "docker run" from crontab

Make an automated backup (every night at 1).

> backup_script.sh

```sh
#!/bin/sh
DB_USERNAME='db_username'
DB_PASSWORD='password'
DB_HOST='db_hostname'
DB_PORT="5432"
DB_NAME='db_name'
BACKUP_DIR='/some/path/backup/'

docker run --rm --name pg-bkup -v $BACKUP_DIR:/backup/ -e "DB_HOST=$DB_HOST" -e "DB_PORT=$DB_PORT" -e "DB_USERNAME=$DB_USERNAME" -e "DB_PASSWORD=$DB_PASSWORD" jkaninda/pg-bkup  bkup -o backup -d $DB_NAME
```

```sh
chmod +x backup_script.sh
```

Your crontab looks like this:

```conf
0 1 * * * /path/to/backup_script.sh
```

## Kubernetes CronJob

Simple Kubernetes CronJob usage:

```yaml
apiVersion: batch/v1
kind: CronJob
metadata:
  name: bkup-job
spec:
  schedule: "0 1 * * *"
  jobTemplate:
    spec:
      template:
        spec:
          backoffLimit: 2
          containers:
          - name: pg-bkup
            image: jkaninda/pg-bkup
            securityContext:
              privileged: true
            command:
            - /bin/sh
            - -c
            - bkup --operation backup -s s3 --path /custom_path
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
              - name: ACCESS_KEY
                value: ""
              - name: SECRET_KEY
                value: ""
              - name: BUCKETNAME
                value: ""
              - name: S3_ENDPOINT
                value: "https://s3.amazonaws.com"
          restartPolicy: Never
```