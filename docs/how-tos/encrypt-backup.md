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
- **Manual Decryption**: Backups encrypted with a GPG public key require manual decryption before restoration.

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
      - DB_USERNAME=username
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

To encrypt backups using a GPG public key, set the `GPG_PUBLIC_KEY` environment variable to the path of your public key file. Backups encrypted with a public key require manual decryption before restoration.

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
      - DB_USERNAME=username
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

If you encrypted your backup using a GPG public key, you must manually decrypt it before restoration. Use the `gnupg` tool for decryption.

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
- **Manual Decryption**: Backups encrypted with a GPG public key require manual decryption using the corresponding private key.
- **Security**: Always keep your GPG passphrase and private key secure. Use Kubernetes Secrets or other secure methods to manage sensitive data.
