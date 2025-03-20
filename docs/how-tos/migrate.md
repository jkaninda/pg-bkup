---
title: Migrate database
layout: default
parent: How Tos
nav_order: 10
---

# Migrate Database

The `migrate` command allows you to transfer a PostgreSQL database from a source to a target database in a single step, combining backup and restore operations.

{: .note }
The `migrate` command eliminates the need for separate backup and restore processes by directly transferring data between databases.

{: .warning }
The migration process is **irreversible**. Always create a backup of your target database before proceeding.

---

## Configuration Steps

1. **Define Source Database**: Provide connection details for the source database.
2. **Define Target Database**: Provide connection details for the target database.
3. **Run Migration**: Execute the `migrate` command to initiate the process.
4. **Migrate All Databases** (Optional): Use the `--all-databases` (`-a`) flag to migrate all databases.

---

## Migrate Database Using Docker CLI

You can also run the migration directly via the Docker CLI.

### 1. Save Environment Variables

Create an environment file (e.g., `your-env`) with your database credentials:

```bash
# Source Database
DB_HOST=postgres
DB_PORT=5432
DB_NAME=database
DB_USERNAME=username
DB_PASSWORD=password

# Target Database
TARGET_DB_HOST=target-postgres
TARGET_DB_PORT=5432
TARGET_DB_NAME=dbname
TARGET_DB_USERNAME=username
TARGET_DB_PASSWORD=password
```

### 2. Run the Migration

Execute the following command:

```bash
docker run --rm --network your_network_name \
  --env-file your-env \
  -v $PWD/backup:/backup/ \
  jkaninda/pg-bkup migrate
```

### 3. Migrate All Databases

To migrate all databases, use:

```bash
docker run --rm --network your_network_name \
  --env-file your-env \
  -v $PWD/backup:/backup/ \
  jkaninda/pg-bkup migrate --all-databases
```

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
