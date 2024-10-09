FROM golang:1.21.0 AS build
WORKDIR /app

# Copy the source code.
COPY . .
# Installs Go dependencies
RUN go mod download

# Build
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/pg-bkup

FROM alpine:3.20.3
ENV DB_HOST=""
ENV DB_NAME=""
ENV DB_USERNAME=""
ENV DB_PASSWORD=""
ENV DB_PORT=5432
ENV STORAGE=local
ENV AWS_S3_ENDPOINT=""
ENV AWS_S3_BUCKET_NAME=""
ENV AWS_ACCESS_KEY=""
ENV AWS_SECRET_KEY=""
ENV AWS_S3_PATH=""
ENV AWS_REGION="us-west-2"
ENV AWS_DISABLE_SSL="false"
ENV AWS_FORCE_PATH_STYLE="true"
ENV GPG_PASSPHRASE=""
ENV SSH_USER=""
ENV SSH_PASSWORD=""
ENV SSH_HOST=""
ENV SSH_IDENTIFY_FILE=""
ENV SSH_PORT=22
ENV REMOTE_PATH=""
ENV FTP_HOST=""
ENV FTP_PORT=21
ENV FTP_USER=""
ENV FTP_PASSWORD=""
ENV TARGET_DB_HOST=""
ENV TARGET_DB_PORT=5432
ENV TARGET_DB_NAME=""
ENV TARGET_DB_USERNAME=""
ENV TARGET_DB_PASSWORD=""
ENV BACKUP_CRON_EXPRESSION=""
ENV TG_TOKEN=""
ENV TG_CHAT_ID=""
ENV TZ=UTC
ARG WORKDIR="/config"
ARG BACKUPDIR="/backup"
ARG BACKUP_TMP_DIR="/tmp/backup"
ARG appVersion="v1.2.11"
ENV VERSION=${appVersion}
LABEL author="Jonas Kaninda"
LABEL version=${appVersion}

RUN apk --update add --no-cache postgresql-client tzdata
RUN mkdir $WORKDIR
RUN mkdir $BACKUPDIR
RUN mkdir -p $BACKUP_TMP_DIR
RUN chmod 777 $WORKDIR
RUN chmod 777 $BACKUPDIR
RUN chmod 777 $BACKUP_TMP_DIR
RUN chmod 777 $WORKDIR
COPY --from=build /app/pg-bkup /usr/local/bin/pg-bkup
RUN chmod +x /usr/local/bin/pg-bkup

RUN ln -s /usr/local/bin/pg-bkup /usr/local/bin/bkup

# Create the backup script and make it executable
RUN echo '#!/bin/sh\n/usr/local/bin/pg-bkup backup "$@"' > /usr/local/bin/backup && \
    chmod +x /usr/local/bin/backup

# Create the restore script and make it executable
RUN echo '#!/bin/sh\n/usr/local/bin/pg-bkup restore "$@"' > /usr/local/bin/restore && \
    chmod +x /usr/local/bin/restore
# Create the migrate script and make it executable
RUN echo '#!/bin/sh\n/usr/local/bin/pg-bkup migrate "$@"' > /usr/local/bin/migrate && \
    chmod +x /usr/local/bin/migrate

WORKDIR $WORKDIR
ENTRYPOINT ["/usr/local/bin/pg-bkup"]
