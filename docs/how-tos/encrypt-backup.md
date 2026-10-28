---
title: Encrypt backups using GPG
layout: default
parent: How Tos
nav_order: 8
---
# Encrypt Backup

The image supports encrypting backups using one of two methods: **GPG with a passphrase** or **GPG with a public key**. When a `GPG_PASSPHRASE` or `GPG_PUBLIC_KEY` environment variable is set, the backup archive will be encrypted and saved as a `.sql.gpg` or `.sql.gz.gpg` file.

{: .warning }
To restore an encrypted backup, you must provide the same GPG passphrase or private key used during the backup process.

---

## Key Features

- **Cipher Algorithm**: `aes256`
- **Automatic Restoration**: Backups encrypted with a GPG passphrase can be restored automatically without manual decryption.
- **Private key restore**: Backups encrypted with a GPG public key are restored by setting `GPG_PRIVATE_KEY` (a path or the inline key) and, if the key is protected, `GPG_PRIVATE_KEY_PASSPHRASE` (defaults to `GPG_PASSPHRASE`). `/config/gnupg/private_key.{asc,gpg}` is detected automatically.

---

## Using GPG Passphrase

To encrypt backups using a GPG passphrase, set the `GPG_PASSPHRASE` environment variable. The backup will be encrypted and can be restored automatically.

### Example Configuration

```yaml
services:
  pg-bkup:
    # In production, lock your image tag to a specific release version
    # instead of using `latest`. Check https://github.com/jkaninda/pg-bkup/releases
    # for available releases.
    image: jkaninda/pg-bkup
    container_name: pg-bkup
    command: backup -d database
    volumes:
      - ./backup:/backup
    environment:
      - DB_PORT=5432
      - DB_HOST=postgres
      - DB_NAME=database
      - DB_USER=username
      - DB_PASSWORD=password
      ## Required to encrypt backup
      - GPG_PASSPHRASE=my-secure-passphrase
    # Ensure the pg-bkup container is connected to the same network as your database
    networks:
      - web

networks:
  web:
```

---

## Using GPG Public Key

To encrypt backups using a GPG public key, set the `GPG_PUBLIC_KEY` environment variable to the path of your public key file. Restore them with the matching private key (`GPG_PRIVATE_KEY`).

### Example Configuration

```yaml
services:
  pg-bkup:
    # In production, lock your image tag to a specific release version
    # instead of using `latest`. Check https://github.com/jkaninda/pg-bkup/releases
    # for available releases.
    image: jkaninda/pg-bkup
    container_name: pg-bkup
    command: backup -d database
    volumes:
      - ./backup:/backup
      - ./public_key.asc:/config/public_key.asc
    environment:
      - DB_PORT=5432
      - DB_HOST=postgres
      - DB_NAME=database
      - DB_USER=username
      - DB_PASSWORD=password
      ## Required to encrypt backup
      - GPG_PUBLIC_KEY=/config/public_key.asc
    # Ensure the pg-bkup container is connected to the same network as your database
    networks:
      - web

networks:
  web:
```

---

## Manual Decryption

Since v2, `restore` decrypts public-key backups itself when `GPG_PRIVATE_KEY` is set. To decrypt a file by hand, for example outside the container, use the `gnupg` tool:

### Decrypt Using a Passphrase

```bash
gpg --batch --passphrase "my-passphrase" \
  --output database_20240730_044201.sql.gz \
  --decrypt database_20240730_044201.sql.gz.gpg
```

### Decrypt Using a Private Key

```bash
gpg --output database_20240730_044201.sql.gz \
  --decrypt database_20240730_044201.sql.gz.gpg
```

---

## Key Notes

- **Automatic Restoration**: Backups encrypted with a GPG passphrase can be restored directly without manual decryption.
- **Private key restore**: Backups encrypted with a GPG public key are restored with the corresponding private key (`GPG_PRIVATE_KEY`).
- **Security**: Always keep your GPG passphrase and private key secure. Use Kubernetes Secrets or other secure methods to manage sensitive data.


## Using age (v2)

[age](https://age-encryption.org) is a simpler alternative to GPG. Set `AGE_RECIPIENTS` to one or more public keys (`age1...` or `ssh-ed25519 ...`, comma or newline separated) or `AGE_RECIPIENTS_FILE`; backups end in `.age`. Restore with `AGE_IDENTITY` (the `AGE-SECRET-KEY-...` line) or `AGE_IDENTITY_FILE` (default `/config/age/identity.txt`). `AGE_PASSPHRASE` encrypts with a passphrase instead.

```yaml
    command: backup -d database
    environment:
      - AGE_RECIPIENTS=age1ql3z7hjy54pw3hyww5ayyfg7zqgvc7w3j2elw8zmrj2kg5sfn9aqmcac8p
```

When both GPG and age keys are set, GPG is used (`ENCRYPTION=age` forces age).
