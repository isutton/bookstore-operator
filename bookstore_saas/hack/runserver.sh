#!/bin/bash

set -x

until nc -zvw10 ${DATABASE_HOST:-postgres} ${DATABASE_PORT:-5432};
do
    echo 'web: waiting for database...';
    sleep 2;
done && python manage.py runserver
