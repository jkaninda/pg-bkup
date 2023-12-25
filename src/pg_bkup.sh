#!/usr/bin/env bash 

set -e
TIME=$(date +%Y%m%d_%H%M%S)
arg0=$(basename "$0" .sh)
blnk=$(echo "$arg0" | sed 's/./ /g')
export OPERATION=backup
export STORAGE=local
export STORAGE_PATH=/backup
export S3_PATH=/pg-bkup
export TIMEOUT=60
export PGPASSWORD=""
export FILE_COMPRESION=true
export CONNECTION=""
export EXECUTION_MODE="default"
export SCHEDULE_PERIOD="0 1 * * *"
export FILE_COMPRESION=true
usage_info()
{
    echo "Usage: \\"
    echo "     $blnk Backup: pg_bkup -o backup -s s3 \\"
    echo "     $blnk Restore: pg_bkup -o restore -s s3 -f my_db.sql \\"
    echo "     $blnk [-o|--operation] [{-f|--file} ] [{-s|--storage} ] [{-h|--help} ] \\"

}
version_info()
{
   echo "Version: $VERSION"
   exit 0
}
usage()
{
    exec 1>2   # Send standard output to standard error
    usage_info
    exit 0
}

error()
{
    echo "$arg0: $*" >&2
    exit 0
}

help()
{
    echo
    echo "  -o |--operation         -- Set operation (default: backup)"
    echo "  -s |--storage           -- Set storage (default: local)"
    echo "  -f |--file              -- Set file name "
    echo "     |--path              -- Set s3 path, without file name"
    echo "  -d |--dbname            -- Set database name "
    echo "  -p |--port              -- Set database port (default: 3306)"
    echo "  -m |--mode              -- Set execution mode (default: default)"
    echo "     |--period            -- Set schedule period time (default: '0 1 * * *')"
    echo "  -t |--timeout           -- Set timeout (default: 120s)"
    echo "  -h |--help              -- Print this help message and exit"
    echo "  -V |--version           -- Print version information and exit"
    exit 0
}

flags()
{
    while test $# -gt 0
    do
        case "$1" in
        (-o|--operation)
            shift
            [ $# = 0 ] && error "No operation specified - restore or backup"
            export OPERATION="$1"
            shift;;
        (-d|--dbname)
            shift
            [ $# = 0 ] && error "No database name specified"
            export DB_NAME="$1"
            shift;;
        (-s|--storage)
            shift
            [ $# = 0 ] && error "No storage specified - local or s3 | default local"
            export STORAGE="$1"
            shift;;
        (-f|--file)
            shift
            [ $# = 0 ] && error "No file specified - file to restore"
            export FILE_NAME="$1"
            shift;;
        (--path)
            shift
            [ $# = 0 ] && error "No s3 path specified - s3 path without file name"
            export S3_PATH="$1"
            shift;;
        (-db|--database)
            shift
            [ $# = 0 ] && error "No database name specified"
            export DB_NAME="$1"
            shift;;
        (-p|--port)
            shift
            [ $# = 0 ] && error "No database name specified"
            export DB_PORT="$1"
            shift;;
        (-m|--mode)
            shift
            [ $# = 0 ] && error "No execution mode specified"
            export EXECUTION_MODE="$1"
            shift;;
        (--period)
            shift
            [ $# = 0 ] && error "No schedule period entered"
            export SCHEDULE_PERIOD="$1"
            shift;;
        (-t|--timeout)
            shift
            [ $# = 0 ] && error "No timeout specified"
            export TIMEOUT="$1"
            shift;;   
        (-h|--help)
            help;;
        (-V|--version)
           version_info;;
        (--)
           help;;
        (*) usage;;
        esac
    done
}
create_pgpass(){
  export CONNECTION=${DB_HOST}:${DB_PORT}:${DB_DATABASE}:${DB_USERNAME}:${DB_PASSWORD}
  echo $CONNECTION > ~/.pgpass
  chmod 600 ~/.pgpass

}
backup()
{
if [[ -z $DB_HOST ]] ||  [[ -z $DB_NAME ]] ||  [[ -z $DB_USERNAME ]] ||  [[ -z $DB_PASSWORD ]]; then
   echo "Please make sure all required options are set "
else
       export PGPASSWORD=${DB_PASSWORD}
       ## Test database connection

       ## Backup database
        pg_dump -h ${DB_HOST} -p ${DB_PORT} -U ${DB_USERNAME} -d ${DB_NAME} -v | gzip > ${STORAGE_PATH}/${DB_NAME}_${TIME}.sql.gz
        echo "$TIME: ${DB_NAME}_${TIME}.sql.gz" | tee -a "${STORAGE_PATH}/history.txt"

        echo "Database has been saved"   
fi
exit 0
}

restore()
{
if [[ -z $DB_HOST ]] ||  [[ -z $DB_NAME ]] ||  [[ -z $DB_USERNAME ]] || [[ -z $DB_PASSWORD ]]; then
   echo "Please make sure all required options are set "
else
    ## Restore database
    export PGPASSWORD=$DB_PASSWORD
     if [ -f "${STORAGE_PATH}/$FILE_NAME" ]; then
         if gzip -t $STORAGE_PATH/$FILE_NAME; then
            zcat ${STORAGE_PATH}/${FILE_NAME} | psql -h ${DB_HOST} -p ${DB_PORT} -U ${DB_USERNAME} -v -d ${DB_NAME}
         else 
            cat ${STORAGE_PATH}/${FILE_NAME} | psql -h ${DB_HOST} -p ${DB_PORT} -U ${DB_USERNAME} -v -d ${DB_NAME}
         fi
        echo "Database has been restored"
      else
        echo "Error, file not found in ${STORAGE_PATH}/${FILE_NAME}"
      fi 
fi
exit
}

s3_backup()
{
 mount_s3
 backup
}

s3_restore()
{
   mount_s3
   restore

}

mount_s3()
{
if [[ -z $ACCESS_KEY ]] ||  [[ -z $SECRET_KEY ]]; then
echo "Please make sure all environment variables are set "
echo "BUCKETNAME=$BUCKETNAME \nACCESS_KEY=$nACCESS_KEY \nSECRET_KEY=$SECRET_KEY"
else
    echo "$ACCESS_KEY:$SECRET_KEY" | tee /etc/passwd-s3fs
    chmod 600 /etc/passwd-s3fs
    echo "Mounting Object storage in /s3mnt .... "
    if [ -z "$(ls -A /s3mnt)" ]; then
       s3fs $BUCKETNAME /s3mnt -o passwd_file=/etc/passwd-s3fs -o use_cache=/tmp/s3cache -o allow_other -o url=$S3_ENDPOINT -o use_path_request_style
       if [ ! -d "/s3mnt$S3_PATH" ]; then
           mkdir -p /s3mnt$S3_PATH
        fi 
    else
     echo "Object storage already mounted in /s3mnt"
    fi
export STORAGE_PATH=/s3mnt$S3_PATH
fi
}
create_crontab_script()
{
TASK=/usr/local/bin/backup_cron.sh
touch $TASK
if [ $STORAGE == 's3' ]
then
cat > "$TASK" <<EOF
#!/usr/bin/env bash 
set -e
bkup --operation backup --dbname $DB_NAME --port $DB_PORT --storage s3 --path $S3_PATH 
EOF
else
cat > "$TASK" <<EOF
#!/usr/bin/env bash 
set -e
bkup --operation backup --dbname $DB_NAME --port $DB_PORT
EOF
fi

chmod +x /usr/local/bin/backup_cron.sh
ln -s /usr/local/bin/backup_cron.sh /usr/local/bin/backup_cron
## Create crontab job
CRON_JOB=/etc/cron.d/backup_cron
touch $CRON_JOB
cat > "$CRON_JOB" <<EOF
$SCHEDULE_PERIOD root exec /bin/bash -c ". /run/supervisord.env; /usr/local/bin/backup_cron.sh >> /var/log/pg-bkup.log"
EOF
chmod 0644 /etc/cron.d/*
crontab /etc/cron.d/backup_cron
}
scheduled_mode()
{
  if [  $OPERATION == 'backup' ]
  then
     create_crontab_script
     echo ""
     echo "**********************************"
     echo "     Starting PostgreSQL Bkup...   "
     echo "***********************************"
     echo "Running in Scheduled mode          "
     echo "Execution period $SCHEDULE_PERIOD"
     echo "Log file in /var/log/pg-bkup.log "
    supervisord -c /etc/supervisor/supervisord.conf
  else
    echo "Scheduled mode supports only backup operation"
    exit 1
  fi
}

flags "$@"
# ?
if [  $EXECUTION_MODE == 'default' ]
then
  if [  $OPERATION != 'backup' ]
  then
     if [ $STORAGE != 's3' ]
     then
          echo "Restore from local"
          restore
      else
        echo "Restore from s3"
        s3_restore
      fi
  else
      if [ $STORAGE != 's3' ]
      then
          echo "Backup to local destination"
          backup
      else
         echo "Backup to s3 storage"
         s3_backup
      fi
   fi
elif [  $EXECUTION_MODE == 'scheduled' ]
then
  scheduled_mode
else
echo "Error, unknow execution mode!"
exit 1
fi