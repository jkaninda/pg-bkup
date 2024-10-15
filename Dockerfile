FROM golang:1.21.0 AS build
WORKDIR /app

# Copy the source code.
COPY . .
# Installs Go dependencies
RUN go mod download

# Build
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/pg-bkup

FROM alpine:3.20.3
ENV TZ=UTC
ARG WORKDIR="/config"
ARG BACKUPDIR="/backup"
ARG BACKUP_TMP_DIR="/tmp/backup"
ARG TEMPLATES_DIR="/config/templates"
ENV VERSION=${appVersion}
LABEL author="Jonas Kaninda"
LABEL version=${appVersion}
LABEL github="github.com/jkaninda/pg-bkup"

RUN apk --update add --no-cache postgresql-client tzdata ca-certificates
RUN mkdir $WORKDIR
RUN mkdir $BACKUPDIR
RUN mkdir $TEMPLATES_DIR
RUN mkdir -p $BACKUP_TMP_DIR
RUN chmod 777 $WORKDIR
RUN chmod 777 $BACKUPDIR
RUN chmod 777 $BACKUP_TMP_DIR
RUN chmod 777 $WORKDIR
COPY --from=build /app/pg-bkup /usr/local/bin/pg-bkup
COPY ./templates/* $TEMPLATES_DIR/
RUN chmod +x /usr/local/bin/pg-bkup

RUN ln -s /usr/local/bin/pg-bkup /usr/local/bin/bkup

# Create the backup script and make it executable
RUN printf '#!/bin/sh\n/usr/local/bin/pg-bkup backup "$@"' > /usr/local/bin/backup && \
    chmod +x /usr/local/bin/backup

# Create the restore script and make it executable
RUN printf '#!/bin/sh\n/usr/local/bin/pg-bkup restore "$@"' > /usr/local/bin/restore && \
    chmod +x /usr/local/bin/restore
# Create the migrate script and make it executable
RUN printf '#!/bin/sh\n/usr/local/bin/pg-bkup migrate "$@"' > /usr/local/bin/migrate && \
    chmod +x /usr/local/bin/migrate

WORKDIR $WORKDIR
ENTRYPOINT ["/usr/local/bin/pg-bkup"]

