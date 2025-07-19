---
title: Migrate database
layout: default
parent: How Tos
nav_order: 10
---

# Migrate PostgreSQL Database

The `migrate` command enables seamless data transfer between PostgreSQL databases. It combines **backup** and **restore** into a single operation — eliminating the need to run them separately.

{: .note }
> The `migrate` command directly transfers data from the source to the target database, simplifying the migration process.

{: .warning }
> **This process is irreversible.** Always back up the **target** database before proceeding.

---

## Configuration Steps

1. **Set Up Source Database**
   Define the connection details for the **source** PostgreSQL database.

2. **Set Up Target Database**
   Define the connection details for the **target** PostgreSQL database.

3. **Run Migration**
   Execute the `migrate` command to begin the data transfer.

4. **Optional: Migrate All Databases**
   Use the `--all-databases` or `-a` flag to migrate **all user databases** from the source server.

5. **Optional: Migrate Entire Instance**
   Use the `--entire-instance` or `-I` flag to migrate the **entire PostgreSQL instance**, including:

    * All databases
    * Roles
    * Tablespaces

{: .note }
> Running a full migration (`--entire-instance`) or bulk database transfer requires **admin privileges** on the PostgreSQL server.

---

## Use Cases

* **Database Migration**: Move data from one PostgreSQL instance to another.
* **Version Upgrades**: Migrate to a newer version of PostgreSQL.
* **Environment Replication**: Clone production to staging or testing environments.

---

## Migrate Using Docker CLI

You can run migrations using Docker by passing environment variables and mounting volumes for intermediate data.

### 1. Create an Environment File

Save your credentials in a `.env` file (e.g., `your-env`):

```bash
# Source Database
DB_HOST=postgres
DB_PORT=5432
DB_NAME=source_db
DB_USERNAME=source_user
DB_PASSWORD=source_pass

# Target Database
TARGET_DB_HOST=target-postgres
TARGET_DB_PORT=5432
TARGET_DB_NAME=target_db
TARGET_DB_USERNAME=target_user
TARGET_DB_PASSWORD=target_pass
```

### 2. Run a Basic Migration

```bash
docker run --rm --network your_network_name \
  --env-file your-env \
  -v $PWD/backup:/backup/ \
  jkaninda/pg-bkup migrate
```

### 3. Migrate All Databases

```bash
docker run --rm --network your_network_name \
  --env-file your-env \
  -v $PWD/backup:/backup/ \
  jkaninda/pg-bkup migrate --all-databases
```

### 4. Migrate the Entire PostgreSQL Instance

```bash
docker run --rm --network your_network_name \
  --env-file your-env \
  -v $PWD/backup:/backup/ \
  jkaninda/pg-bkup migrate --entire-instance
```

---

### Example: Docker Compose Configuration

Below is an example `docker-compose.yml` configuration for migrating a PostgreSQL database:

```yaml
services:
  pg-bkup:
    # Use a specific version instead of `latest` in production.
    # Check available releases at: https://github.com/jkaninda/pg-bkup/releases
    image: jkaninda/pg-bkup
    container_name: pg-bkup
    command: migrate
    volumes:
      - ./backup:/backup
    environment:
      # Source Database Configuration
      - DB_HOST=postgres
      - DB_PORT=5432
      - DB_NAME=database
      - DB_USERNAME=username
      - DB_PASSWORD=password
      # Optional: Use JDBC connection string
      #- DB_URL=jdbc:postgresql://postgres:5432/database?user=username&password=password

      # Target Database Configuration
      - TARGET_DB_HOST=target-postgres
      - TARGET_DB_PORT=5432
      - TARGET_DB_NAME=dbname
      - TARGET_DB_USERNAME=username
      - TARGET_DB_PASSWORD=password
      # Optional: Use JDBC connection string
      #- TARGET_DB_URL=jdbc:postgresql://target-postgres:5432/dbname?user=username&password=password

    # Ensure the `pg-bkup` container is on the same network as your databases
    networks:
      - web

networks:
  web:
```
