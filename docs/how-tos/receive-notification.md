---
title: Receive notifications
layout: default
parent: How Tos
nav_order: 13
---

# Receive Notifications

You can configure the system to send email (SMTP or Posta) or Telegram notifications when a backup succeeds or fails. 

This section explains how to set up and customize notifications.

---

## Email Notifications

To send email notifications, provide SMTP credentials, a sender address, and recipient addresses. Notifications will be sent for both successful and failed backup runs.

### Example: Email Notification Configuration

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
      - DB_USER=username
      - DB_PASSWORD=password
      ## SMTP Configuration
      - MAIL_HOST=smtp.example.com
      - MAIL_PORT=587
      - MAIL_USERNAME=your-email@example.com
      - MAIL_PASSWORD=your-email-password
      - MAIL_FROM=Backup Jobs <backup@example.com>
      ## Multiple recipients separated by a comma
      - MAIL_TO=me@example.com,team@example.com,manager@example.com
      ## STARTTLS (default), tls or none; set MAIL_INSECURE_SKIP_VERIFY=true only for self-signed certificates
      - MAIL_TLS=starttls
      - MAIL_INSECURE_SKIP_VERIFY=false
      ## Time format for notifications
      - TIME_FORMAT=2006-01-02 at 15:04:05
      ## Backup reference (e.g., database/cluster name or server name)
      - BACKUP_REFERENCE=database/Paris cluster
    networks:
      - web

networks:
  web:
```

---

## Posta Notifications

[Posta](https://github.com/goposta/posta) can deliver the notification emails instead of an SMTP server. Set the Posta URL, an API key with the `send` scope and the recipients. By default the same HTML email is sent (with a plain-text version). To use a template stored in Posta instead, set `POSTA_TEMPLATE` (and optionally `POSTA_TEMPLATE_ERROR` for failures); the notification fields (`database`, `file`, `duration`, `storage`, `location`, `size`, `backup_size`, `sha256`, `encryption`, `error`, `time`, `success`, `tool`...) are passed as template data.

### Example: Posta Notification Configuration

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
      - DB_HOST=db
      - DB_NAME=database
      - DB_USER=username
      - DB_PASSWORD=password
      ## Posta
      - POSTA_URL=https://posta.example.com
      - POSTA_API_KEY=psk_your_api_key
      - POSTA_FROM=Backup Jobs <backup@example.com>
      ## Multiple recipients separated by a comma (default MAIL_TO)
      - POSTA_TO=me@example.com,team@example.com
      ## Optional: templates stored in Posta
      #- POSTA_TEMPLATE=backup-success
      #- POSTA_TEMPLATE_ERROR=backup-failure
      #- POSTA_LANGUAGE=en
      ## Only for account-wide API keys
      #- POSTA_WORKSPACE_ID=42
      ## Only notify failures
      #- NOTIFY_ON=failure
```

---

## Telegram Notifications

To send Telegram notifications, provide your bot token and chat ID. Notifications will be sent for both successful and failed backup runs.

### Example: Telegram Notification Configuration

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
      - DB_USER=username
      - DB_PASSWORD=password
      ## Telegram Configuration
      - TG_TOKEN=[BOT ID]:[BOT TOKEN]
      - TG_CHAT_ID=your-chat-id
      ## Time format for notifications
      - TIME_FORMAT=2006-01-02 at 15:04:05
      ## Backup reference (e.g., database/cluster name or server name)
      - BACKUP_REFERENCE=database/Paris cluster
    networks:
      - web

networks:
  web:
```

---

## Customize Notifications

You can customize the title and body of notifications using Go templates. Template files must be mounted inside the container at `/config/templates`. The following templates are supported:

- `email.tmpl`: Template for successful email notifications.
- `telegram.tmpl` (or `text.tmpl`): Template for successful Telegram, Slack, Discord and ntfy notifications.
- `email-error.tmpl`: Template for failed email notifications.
- `telegram-error.tmpl` (or `text-error.tmpl`): Template for failed Telegram, Slack, Discord and ntfy notifications.

### Template Data

The following data is passed to the templates:

- `Database`: Database name.
- `StartTime`: Backup start time.
- `EndTime`: Backup end time.
- `Storage`: Backup storage type (e.g., local, S3, SSH).
- `BackupLocation`: Backup file location.
- `BackupSize`: Backup file size in bytes.
- `BackupReference`: Backup reference (e.g., database/cluster name or server name).
- `Error`: Error message (only for error templates).

---

### Example Templates

#### `email.tmpl` (Successful Backup)

```html
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
    <li>Backup Reference: {{.BackupReference}}</li>
</ul>
<p>Best regards,</p>
```

#### `telegram.tmpl` (Successful Backup)

```html
✅ Database Backup Notification – {{.Database}}
Hi,
Backup of the {{.Database}} database has been successfully completed on {{.EndTime}}.

Backup Details:
- Database Name: {{.Database}}
- Backup Start Time: {{.StartTime}}
- Backup End Time: {{.EndTime}}
- Backup Storage: {{.Storage}}
- Backup Location: {{.BackupLocation}}
- Backup Size: {{.BackupSize}} bytes
- Backup Reference: {{.BackupReference}}
```

#### `email-error.tmpl` (Failed Backup)

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
    <li>Backup Reference: {{.BackupReference}}</li>
</ul>
</body>
</html>
```

#### `telegram-error.tmpl` (Failed Backup)

```html
🔴 Urgent: Database Backup Failure Notification

An error occurred during database backup.
Failure Details:

Error Message: {{.Error}}
Date: {{.EndTime}}
Backup Reference: {{.BackupReference}}
```

---

## Key Notes

- **SMTP Configuration**: `MAIL_TLS` selects `starttls` (default), `tls` or `none`; certificates are verified unless `MAIL_INSECURE_SKIP_VERIFY=true`. The 1.x `MAIL_SKIP_TLS` is still read, now with its documented meaning (`true` skips verification).
- **Other channels**: Posta (`POSTA_*`), Slack, Discord, ntfy, a generic webhook and healthchecks.io pings are also supported (`SLACK_WEBHOOK_URL`, `DISCORD_WEBHOOK_URL`, `NTFY_URL`, `WEBHOOK_URL`, `HEALTHCHECK_PING_URL`); see the [configuration reference](../reference/).
- **Telegram Configuration**: Obtain your bot token and chat ID from Telegram.
- **Custom Templates**: Mount custom templates to `/config/templates` to override default notifications.
- **Time Format**: Use the `TIME_FORMAT` environment variable to customize the timestamp format in notifications.