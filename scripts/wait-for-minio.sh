#!/bin/sh

set -e

host="$1"
shift
cmd="$@"

#until PGPASSWORD=$POSTGRES_PASSWORD psql -h "$host" -U "postgres" -c '\q'; do
until PGPASSWORD=$POSTGRES_PASSWORD psql -h "$host" -U "postgres" -c '\q'; do
  >&2 echo "Minio is unavailable - sleeping"
  sleep 1
done

>&2 echo "Minio is up - executing command"
exec $cmd
