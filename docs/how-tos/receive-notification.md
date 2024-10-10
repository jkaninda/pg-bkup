---
title: Receive notifications
layout: default
parent: How Tos
nav_order: 12
---
Send Email or Telegram notifications on successfully or failed backup.

### Email
To send out email notifications on failed or successfully backup runs, provide SMTP credentials, a sender and a recipient:

```yaml
services:
  pg-bkup:
    image: jkaninda/pg-bkup
    container_name: pg-bkup
    command: backup
    volumes:
      - ./backup:/backup
    environment:
      - DB_PORT=5432
      - DB_HOST=postgres
      - DB_NAME=database
      - DB_USERNAME=username
      - DB_PASSWORD=password
      - MAIL_HOST=
      - MAIL_PORT=587
      - MAIL_USERNAME=
      - MAIL_PASSWORD=!
      - MAIL_FROM=
      - MAIL_TO=me@example.com,team@example.com,manager@example.com
      - MAIL_SKIP_TLS=false
      ## Time format for notification 
      - TIME_FORMAT=2006-01-02 at 15:04:05
      ## Backup reference, in case you want to identify every backup instance
      - BACKUP_REFERENCE=database/Paris cluster
    networks:
      - web
networks:
  web:
```

### Telegram

```yaml
services:
  pg-bkup:
    image: jkaninda/pg-bkup
    container_name: pg-bkup
    command: backup
    volumes:
      - ./backup:/backup
    environment:
      - DB_PORT=5432
      - DB_HOST=postgres
      - DB_NAME=database
      - DB_USERNAME=username
      - DB_PASSWORD=password
      - TG_TOKEN=[BOT ID]:[BOT TOKEN]
      - TG_CHAT_ID=
      ## Time format for notification 
      - TIME_FORMAT=2006-01-02 at 15:04:05
      ## Backup reference, in case you want to identify every backup instance
      - BACKUP_REFERENCE=database/Paris cluster
    networks:
      - web
networks:
  web:
```

### Customize notifications

The title and body of the notifications can be tailored to your needs using Go templates.
Template sources must be mounted inside the container in /config/templates:

- email.template: Email notification template
- telegram.template: Telegram notification template
- email-error.template: Error notification template
- telegram-error.template: Error notification template

### Data

Here is a list of all data passed to the template:
- `Database` : Database name
- `StartTime`: Backup start time process
- `EndTime`: Backup start time process
- `Storage`: Backup storage
- `BackupLocation`: Backup location
- `BackupSize`: Backup size
- `BackupReference`: Backup reference(eg: database/cluster name or server name)

>  email.template:


```html
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>✅  Database Backup Notification – {{.Database}}</title>
</head>
<body>
<h2>Hi,</h2>
<p>Backup of the {{.Database}} database has been successfully completed on {{.EndTime}}.</p>
<h3>Backup Details:</h3>
<ul>
    <li>Database Name: {{.Database}}</li>
    <li>Backup Start Time: {{.StartTime}}</li>
    <li>Backup End Time: {{.EndTime}}</li>
    <li>Backup Storage: {{.Storage}}</li>
    <li>Backup Location: {{.BackupLocation}}</li>
    <li>Backup Size: {{.BackupSize}} bytes</li>
    <li>Backup Reference: {{.BackupReference}} </li>
</ul>
<p>Best regards,</p>
</body>
</html>
```

> telegram.template

```html
✅  Database Backup Notification – {{.Database}}
Hi,
Backup of the {{.Database}} database has been successfully completed on {{.EndTime}}.

Backup Details:
- Database Name: {{.Database}}
- Backup Start Time: {{.StartTime}}
- Backup EndTime: {{.EndTime}}
- Backup Storage: {{.Storage}}
- Backup Location: {{.BackupLocation}}
- Backup Size: {{.BackupSize}} bytes
- Backup Reference: {{.BackupReference}}
```

> email-error.template

```html
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>🔴 Urgent: Database Backup Failure Notification</title>
</head>
<body>
<h2>Hi,</h2>
<p>An error occurred during database backup.</p>
<h3>Failure Details:</h3>
<ul>
<li>Error Message: {{.Error}}</li>
<li>Date: {{.EndTime}}</li>
<li>Backup Reference: {{.BackupReference}} </li>
</ul>
</body>
</html>
```

> telegram-error.template


```html
🔴 Urgent: Database Backup Failure Notification

An error occurred during database backup.
Failure Details:

Error Message: {{.Error}}
Date: {{.EndTime}}
```