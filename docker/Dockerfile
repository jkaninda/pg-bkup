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
ENV BUCKET_NAME=""
ENV ACCESS_KEY=""
ENV SECRET_KEY=""
ENV REGION=""
ENV SSH_USER=""
ENV SSH_PASSWORD=""
ENV SSH_HOST_NAME=""
ENV SSH_IDENTIFY_FILE="/root/.ssh/id_rsa"
ENV GPG_PASS_PHRASE=""
ENV SSH_PORT="22"
ENV S3_ENDPOINT=https://s3.amazonaws.com
ARG DEBIAN_FRONTEND=noninteractive
ENV VERSION="v0.6"
LABEL authors="Jonas Kaninda"

RUN apt-get update -qq

RUN apt install postgresql-client postgresql-client-common supervisor cron openssh-client gnupg -y

# Clear cache
RUN apt-get clean && rm -rf /var/lib/apt/lists/*

RUN mkdir /s3mnt
RUN mkdir /tmp/s3cache
RUN chmod 777 /s3mnt
RUN chmod 777 /tmp/s3cache

COPY --from=build /app/pg-bkup /usr/local/bin/pg-bkup
RUN chmod +x /usr/local/bin/pg-bkup

RUN ln -s /usr/local/bin/pg-bkup /usr/local/bin/bkup

ADD docker/supervisord.conf /etc/supervisor/supervisord.conf


RUN mkdir /backup
RUN mkdir /tmp/backup
WORKDIR /root
