FROM ubuntu:24.04
ENV DB_HOST=""
ENV DB_NAME=""
ENV DB_USERNAME=""
ENV DB_PASSWORD=""
ENV DB_PORT="5432"
ENV DESTINATION=local
ENV STORAGE=local
ENV SOURCE=local
ENV BUCKETNAME=""
ENV ACCESS_KEY=""
ENV SECRET_KEY=""
ENV S3_ENDPOINT=https://s3.amazonaws.com
ARG DEBIAN_FRONTEND=noninteractive
ENV VERSION="0.2"

RUN apt-get update -qq
RUN apt-get install build-essential libcurl4-openssl-dev libxml2-dev mime-support -y
RUN apt install s3fs postgresql-client postgresql-client-common libpq-dev supervisor cron -y

# Clear cache
RUN apt-get clean && rm -rf /var/lib/apt/lists/*

RUN mkdir /s3mnt
RUN mkdir /tmp/s3cache
RUN chmod 777 /s3mnt
RUN chmod 777 /tmp/s3cache

COPY src/pg_bkup.sh /usr/local/bin/
RUN chmod +x /usr/local/bin/pg_bkup.sh

ADD src/supervisord.conf /etc/supervisor/supervisord.conf

RUN ln -s /usr/local/bin/pg_bkup.sh /usr/local/bin/pg_bkup
RUN ln -s /usr/local/bin/pg_bkup.sh /usr/local/bin/bkup

RUN mkdir /backup
WORKDIR /backup