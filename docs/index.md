---
title: Overview
layout: home
nav_order: 1
---

# About PG-BKUP
{:.no_toc}

**PG-BKUP** is a Docker container image designed to **backup, restore, and migrate PostgreSQL databases**.
It supports a variety of storage options and ensures data security through GPG encryption.

## Features

- **Storage Options:**
    - Local storage
    - AWS S3 or any S3-compatible object storage
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

## Use Cases

- **Automated Recurring Backups:** Schedule regular backups for PostgreSQL databases.
- **Cross-Environment Migration:** Easily migrate your PostgreSQL databases across different environments using supported storage options.
- **Secure Backup Management:** Protect your data with GPG encryption.


We are open to receiving stars, PRs, and issues!


{: .fs-6 .fw-300 }

---

{: .note }
Code and documentation for `v1` version on [this branch][v1-branch].

[v1-branch]: https://github.com/jkaninda/pg-bkup

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

We decided to publish this image as a simpler and more lightweight alternative because of the following requirements:

- The original image is based on `Alpine` and requires additional tools, making it heavy.
- This image is written in Go.
- `arm64` and `arm/v7` architectures are supported.
- Docker in Swarm mode is supported.
- Kubernetes is supported.
