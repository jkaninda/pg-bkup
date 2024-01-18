#!/usr/bin/env bash
if [ $# -eq 0 ]
  then
    tag='latest'
  else
    tag=$1
fi

#go build
#CGO_ENABLED=0 GOOS=linux go build

docker build -f docker/Dockerfile  -t jkaninda/pg-bkup:$tag .

#docker compose up -d --force-recreate