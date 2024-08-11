# PostgreSQL Backup
PostgreSQL Backup is a Docker container image that can be used to backup and restore Postgres database. It supports local storage, AWS S3 or any S3 Alternatives for Object Storage, and SSH compatible storage.
It also supports __encrypting__ your backups using GPG.

The [jkaninda/pg-bkup](https://hub.docker.com/r/jkaninda/pg-bkup) Docker image can be deployed on Docker, Docker Swarm and Kubernetes.
It handles __recurring__ backups of postgres database on Docker and can be deployed as __CronJob on Kubernetes__ using local, AWS S3 or SSH compatible storage.

It also supports __encrypting__ your backups using GPG.

[![Build](https://github.com/jkaninda/pg-bkup/actions/workflows/release.yml/badge.svg)](https://github.com/jkaninda/pg-bkup/actions/workflows/release.yml)
[![Go Report](https://goreportcard.com/badge/github.com/jkaninda/mysql-bkup)](https://goreportcard.com/report/github.com/jkaninda/pg-bkup)
![Docker Image Size (latest by date)](https://img.shields.io/docker/image-size/jkaninda/pg-bkup?style=flat-square)
![Docker Pulls](https://img.shields.io/docker/pulls/jkaninda/pg-bkup?style=flat-square)

- Docker
- Docker Swarm
- Kubernetes

## Documentation is found at <https://jkaninda.github.io/pg-bkup>


## Links:

- [Docker Hub](https://hub.docker.com/r/jkaninda/pg-bkup)
- [Github](https://github.com/jkaninda/pg-bkup)

## MySQL solution :

- [MySQL](https://github.com/jkaninda/mysql-bkup)

## Storage:
- Local
- AWS S3 or any S3 Alternatives for Object Storage
- SSH

## Quickstart

### Simple backup using Docker CLI

To run a one time backup, bind your local volume to `/backup` in the container and run the `backup` command:

```shell
 docker run --rm --network your_network_name \
 -v $PWD/backup:/backup/ \
 -e "DB_HOST=dbhost" \
 -e "DB_USERNAME=username" \
 -e "DB_PASSWORD=password" \
 jkaninda/pg-bkup backup -d database_name
```

Alternatively, pass a `--env-file` in order to use a full config as described below.

```yaml
 docker run --rm --network your_network_name \
 --env-file your-env-file
 -v $PWD/backup:/backup/ \
 jkaninda/pg-bkup backup -d database_name
```

### Simple backup in docker compose file

```yaml
services:
  pg-bkup:
    # In production, it is advised to lock your image tag to a proper
    # release version instead of using `latest`.
    # Check https://github.com/jkaninda/pg-bkup/releases
    # for a list of available releases.
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
    # pg-bkup container must be connected to the same network with your database
    networks:
       - web
networks:
  web:
```
## Deploy on Kubernetes

For Kubernetes, you don't need to run it in scheduled mode. You can deploy it as CronJob.

### Simple Kubernetes CronJob usage:

```yaml
apiVersion: batch/v1
kind: CronJob
metadata:
  name: backup-job
spec:
  schedule: "0 1 * * *"
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
              - s3
              - --disable-compression
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
## Available image registries

This Docker image is published to both Docker Hub and the GitHub container registry.
Depending on your preferences and needs, you can reference both `jkaninda/pg-bkup` as well as `ghcr.io/jkaninda/pg-bkup`:

```
docker pull jkaninda/pg-bkup
docker pull ghcr.io/jkaninda/pg-bkup
```

Documentation references Docker Hub, but all examples will work using ghcr.io just as well.

## Supported Engines

This image is developed and tested against the Docker CE engine and Kubernetes exclusively.
While it may work against different implementations, there are no guarantees about support for non-Docker engines.

## References

We decided to publish this image as a simpler and more lightweight alternative because of the following requirements:

- The original image is based on `ubuntu` and requires additional tools, making it heavy.
- This image is written in Go.
- `arm64` and `arm/v7` architectures are supported.
- Docker in Swarm mode is supported.
- Kubernetes is supported.


## License

This project is licensed under the MIT License. See the LICENSE file for details.

## Authors

**Jonas Kaninda**
- <https://github.com/jkaninda>

## Copyright

Copyright (c) [2023] [Jonas Kaninda]
