---
title: Backup all databases in the server
layout: default
parent: How Tos
nav_order: 12
---

# Backup All Databases

PG-Bkup supports backing up all databases on the server using the `--all-databases` (`-a`) flag. By default, this creates separate backup files for each database. If you prefer a single backup file, you can use the `--all-in-one` (`-A`) flag.

Backing up all databases is useful for creating a snapshot of the entire database server, whether for disaster recovery or migration purposes.
## Backup Modes

### Separate Backup Files (Default)

Using --all-databases without --all-in-one creates individual backup files for each database.

- Creates separate backup files for each database.
- Provides more flexibility in restoring individual databases or tables.
- Can be more manageable in cases where different databases have different retention policies.
- Might take slightly longer due to multiple file operations.
- It is the default behavior when using the `--all-databases` flag.
- It does not backup system databases (`postgres`,`template0`, `template1`,...).

**Command:**

```bash
docker run --rm --network your_network_name \
  -v $PWD/backup:/backup/ \
  -e "DB_HOST=dbhost" \
  -e "DB_PORT=5432" \
  -e "DB_USERNAME=username" \
  -e "DB_PASSWORD=password" \
  jkaninda/pg-bkup backup --all-databases
```
### Single Backup File

Using --all-in-one (-A) creates a single backup file containing all databases.

- Creates a single backup file containing all databases.
- Easier to manage if you need to restore everything at once.
- Faster to back up and restore in bulk.
- Can be problematic if you only need to restore a specific database or table.
- It is recommended to use this option for disaster recovery purposes.
- It backups system databases as well.

```bash
docker run --rm --network your_network_name \
  -v $PWD/backup:/backup/ \
  -e "DB_HOST=dbhost" \
  -e "DB_PORT=5432" \
  -e "DB_USERNAME=username" \
  -e "DB_PASSWORD=password" \
  jkaninda/pg-bkup backup --all-in-one
```

### When to Use Which?

- Use `--all-in-one` if you want a quick, simple backup for disaster recovery where you'll restore everything at once.
- Use `--all-databases` if you need granularity in restoring specific databases or tables without affecting others.
