# syntax=docker/dockerfile:1
FROM --platform=$BUILDPLATFORM golang:1.26.3-alpine3.23 AS build
WORKDIR /src
ARG TARGETOS
ARG TARGETARCH
ARG TARGETVARIANT
ARG appVersion="dev"
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN GOARM="${TARGETVARIANT#v}" CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build -trimpath -ldflags="-s -w -X main.Version=${appVersion}" -o /out/pg-bkup .

FROM alpine:3.23.4
ARG appVersion="dev"
ENV TZ=UTC
LABEL author="Jonas Kaninda" \
      version=${appVersion} \
      org.opencontainers.image.source="https://github.com/jkaninda/pg-bkup" \
      org.opencontainers.image.version=${appVersion}


RUN apk add --no-cache \
      postgresql16-client=16.15-r0 \
      postgresql17-client=17.11-r0 \
      postgresql18-client=18.6-r0 \
      tini=0.19.0-r3 \
      tzdata ca-certificates && \
    mkdir -p /config /backup /tmp/backup /config/templates && \
    chmod 0777 /backup && chmod 1777 /tmp/backup

COPY --from=build /out/pg-bkup /usr/local/bin/pg-bkup
RUN ln -s /usr/local/bin/pg-bkup /usr/local/bin/bkup && \
    printf '#!/bin/sh\nexec /usr/local/bin/pg-bkup backup "$@"\n' > /usr/local/bin/backup && \
    printf '#!/bin/sh\nexec /usr/local/bin/pg-bkup restore "$@"\n' > /usr/local/bin/restore && \
    printf '#!/bin/sh\nexec /usr/local/bin/pg-bkup migrate "$@"\n' > /usr/local/bin/migrate && \
    chmod +x /usr/local/bin/backup /usr/local/bin/restore /usr/local/bin/migrate

WORKDIR /config

HEALTHCHECK --interval=1m --timeout=10s --start-period=1m CMD ["/usr/local/bin/pg-bkup", "health"]
ENTRYPOINT ["/sbin/tini", "--", "/usr/local/bin/pg-bkup"]
