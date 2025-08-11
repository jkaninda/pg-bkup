FROM golang:1.24.6 AS build
WORKDIR /app
ARG appVersion=""
# Copy the source code.
COPY . .
# Installs Go dependencies
RUN go mod download

# Build
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-X 'github.com/jkaninda/pg-bkup/utils.Version=${appVersion}'" -o /app/pg-bkup

FROM alpine:3.22.1
ENV TZ=UTC
ARG WORKDIR="/config"
ARG BACKUPDIR="/backup"
ARG BACKUP_TMP_DIR="/tmp/backup"
ARG TEMPLATES_DIR="/config/templates"
ARG appVersion=""
ENV VERSION=${appVersion}
LABEL author="Jonas Kaninda"
LABEL version=${appVersion}
LABEL github="github.com/jkaninda/pg-bkup"

RUN apk --update add --no-cache postgresql-client tzdata ca-certificates
RUN mkdir -p $WORKDIR $BACKUPDIR $TEMPLATES_DIR $BACKUP_TMP_DIR && \
     chmod a+rw $WORKDIR $BACKUPDIR $BACKUP_TMP_DIR
COPY --from=build /app/pg-bkup /usr/local/bin/pg-bkup
COPY ./templates/* $TEMPLATES_DIR/
RUN chmod +x /usr/local/bin/pg-bkup && \
    ln -s /usr/local/bin/pg-bkup /usr/local/bin/bkup

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

