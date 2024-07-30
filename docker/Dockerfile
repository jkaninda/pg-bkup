FROM golang:1.21.0 AS build
WORKDIR /app

# Copy the source code.
COPY . .
# Installs Go dependencies
RUN go mod download

# Build
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/pg-bkup

FROM ubuntu:24.04
ENV DB_HOST=""
ENV DB_NAME=""
ENV DB_USERNAME=""
ENV DB_PASSWORD=""
ENV DB_PORT="5432"
ENV STORAGE=local
ENV AWS_S3_ENDPOINT=""
ENV AWS_S3_BUCKET_NAME=""
ENV AWS_ACCESS_KEY=""
ENV AWS_SECRET_KEY=""
ENV AWS_REGION="us-west-2"
ENV AWS_DISABLE_SSL="false"
ENV GPG_PASSPHRASE=""
ENV SSH_USER=""
ENV SSH_REMOTE_PATH=""
ENV SSH_PASSWORD=""
ENV SSH_HOST_NAME=""
ENV SSH_IDENTIFY_FILE=""
ENV SSH_PORT="22"
ARG DEBIAN_FRONTEND=noninteractive
ENV VERSION="v1.0"
ARG WORKDIR="/app"
ARG BACKUPDIR="/backup"
ARG BACKUP_TMP_DIR="/tmp/backup"
ARG BACKUP_CRON="/etc/cron.d/backup_cron"
ARG BACKUP_CRON_SCRIPT="/usr/local/bin/backup_cron.sh"
LABEL author="Jonas Kaninda"

RUN apt-get update -qq

RUN apt install postgresql-client postgresql-client-common supervisor cron openssh-client gnupg -y

# Clear cache
RUN apt-get clean && rm -rf /var/lib/apt/lists/*

RUN mkdir $WORKDIR
RUN mkdir $BACKUPDIR
RUN mkdir -p $BACKUP_TMP_DIR
RUN chmod 777 $WORKDIR
RUN chmod 777 $BACKUPDIR
RUN chmod 777 $BACKUP_TMP_DIR
RUN touch $BACKUP_CRON && \
    touch $BACKUP_CRON_SCRIPT && \
    chmod 777 $BACKUP_CRON && \
    chmod 777 $BACKUP_CRON_SCRIPT

COPY --from=build /app/pg-bkup /usr/local/bin/pg-bkup
RUN chmod +x /usr/local/bin/pg-bkup

RUN ln -s /usr/local/bin/pg-bkup /usr/local/bin/bkup

ADD docker/supervisord.conf /etc/supervisor/supervisord.conf

WORKDIR $WORKDIR