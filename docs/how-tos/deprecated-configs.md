---
title: Update deprecated configurations
layout: default
parent: How Tos
nav_order: 14
---

# Update deprecated configurations

Deprecated names keep working in v2, and pg-bkup logs a warning the first time one is used, for example:

```
level=WARN msg="FTP_HOST_NAME is deprecated, use FTP_HOST"
```

Run `pg-bkup config` to list every resolved setting with the name it was read from. The full table of renamed variables and behaviour changes is in [Upgrading to v2](../../upgrade/).
