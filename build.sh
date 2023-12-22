#!/usr/bin/env bash
if [ $# -eq 0 ]
  then
    tag='latest'
  else
    tag=$1
fi

docker build -f src/docker/Dockerfile  -t jkaninda/pg-bkup:$tag .

docker-compose up -d