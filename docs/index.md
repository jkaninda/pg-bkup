---
title: Overview
layout: home
nav_order: 1
---

# About pg-bkup
{:.no_toc}
pg-bkup is a Docker container image that can be used to backup and restore Postgres database. It supports local storage, AWS S3 or any S3 Alternatives for Object Storage, and SSH compatible storage.
It also supports __encrypting__ your backups using GPG.

We are open to receiving stars, PRs, and issues!


{: .fs-6 .fw-300 }

---

The [jkaninda/pg-bkup](https://hub.docker.com/r/jkaninda/pg-bkup) Docker image can be deployed on Docker, Docker Swarm and Kubernetes. 
It handles __recurring__ backups of postgres database on Docker and can be deployed as __CronJob on Kubernetes__ using local, AWS S3 or SSH compatible storage.

It also supports __encrypting__ your backups using GPG.

{: .note }
Code and documentation for `v1` version on [this branch][v1-branch].

[v1-branch]: https://github.com/jkaninda/pg-bkup

---

## Quickstart

### Simple backup using Docker CLI

To run a one time backup, bind your local volume to `/backup` in the container and run the `pg-bkup backup` command:

```shell
 docker run --rm --network your_network_name \
 -v $PWD/backup:/backup/ \
 -e "DB_HOST=dbhost" \
 -e "DB_USERNAME=username" \
 -e "DB_PASSWORD=password" \
 jkaninda/pg-bkup  pg-bkup backup -d database_name
```

Alternatively, pass a `--env-file` in order to use a full config as described below.

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
    command:
      - /bin/sh
      - -c
      - pg-bkup backup
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

## Available image registries

This Docker image is published to both Docker Hub and the GitHub container registry.
Depending on your preferences and needs, you can reference both `jkaninda/pg-bkup` as well as `ghcr.io/jkaninda/pg-bkup`:

```
docker pull jkaninda/pg-bkup:v1.2.0
docker pull ghcr.io/jkaninda/pg-bkup:v1.2.0
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
