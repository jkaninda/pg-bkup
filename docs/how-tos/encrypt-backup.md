---
title: Encrypt backups using GPG
layout: default
parent: How Tos
nav_order: 8
---
# Encrypt backup

The image supports encrypting backups using one of two available methods: GPG with passphrase or GPG with a public key.


The image supports encrypting backups using GPG out of the box. In case a `GPG_PASSPHRASE` or `GPG_PUBLIC_KEY` environment variable is set, the backup archive will be encrypted using the given key and saved as a sql.gpg file instead or sql.gz.gpg.
Suppose you used a GPG public key during the backup process. In that case, you need to decrypt your backup before restoration because decryption using a `GPG private` key is not fully supported.


{: .warning }
To restore an encrypted backup, you need to provide the same GPG passphrase used during backup process.

- GPG home directory `/config/gnupg`
- Cipher algorithm `aes256`

{: .note }
The backup encrypted using `GPG passphrase` method can be restored automatically, no need to decrypt it before restoration.  


To decrypt manually, you need to install `gnupg`

```shell
gpg --batch --passphrase "my-passphrase" \
--output database_20240730_044201.sql.gz \
--decrypt database_20240730_044201.sql.gz.gpg
```
Using your private key
```shell
gpg --output database_20240730_044201.sql.gz --decrypt database_20240730_044201.sql.gz.gpg
```
## Using GPG passphrase

```yml
services:
  pg-bkup:
    # In production, it is advised to lock your image tag to a proper
    # release version instead of using `latest`.
    # Check https://github.com/jkaninda/pg-bkup/releases
    # for a list of available releases.
    image: jkaninda/pg-bkup
    container_name: pg-bkup
    command: backup -d database
    volumes:
      - ./backup:/backup
    environment:
      - DB_PORT=5432
      - DB_HOST=postgres
      - DB_NAME=database
      - DB_USERNAME=username
      - DB_PASSWORD=password
      ## Required to encrypt backup
      - GPG_PASSPHRASE=my-secure-passphrase
    # pg-bkup container must be connected to the same network with your database
    networks:
      - web
networks:
  web:
```

## Using GPG Public Key

```yml
services:
  pg-bkup:
    # In production, it is advised to lock your image tag to a proper
    # release version instead of using `latest`.
    # Check https://github.com/jkaninda/pg-bkup/releases
    # for a list of available releases.
    image: jkaninda/pg-bkup
    container_name: pg-bkup
    command: backup -d database
    volumes:
      - ./backup:/backup
    environment:
      - DB_PORT=5432
      - DB_HOST=postgres
      - DB_NAME=database
      - DB_USERNAME=username
      - DB_PASSWORD=password
      ## Required to encrypt backup
      - GPG_PUBLIC_KEY=/config/public_key.asc
    # pg-bkup container must be connected to the same network with your database
    networks:
      - web
networks:
  web:
```
