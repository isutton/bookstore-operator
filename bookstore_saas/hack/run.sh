#!/bin/bash

# This script exists to aid the software wiring, probably could be substituted by podman-compose as
# it is doing a very similar thing but much poorly it seems.

set -x

DATABASE_USER=admin
DATABASE_PASS=Passw0rd
DATABASE_NAME=bookstore
DATABASE_HOST=postgres
DATABASE_PORT=5432

NETWORK_NAME=bookstore

podman network create $NETWORK_NAME

podman pod create			\
       --name bookstore-database	\
       --network $NETWORK_NAME		\
       -p $DATABASE_PORT:$DATABASE_PORT 

podman run -d					\
       --name postgres				\
       --network $NETWORK_NAME			\
       --pod=bookstore-database			\
       -e POSTGRES_USER=$DATABASE_USER		\
       -e POSTGRES_PASSWORD=$DATABASE_PASS	\
       -e POSTGRES_DB=$DATABASE_NAME		\
       postgres:14 

podman run -d --rm -it			\
       --name web			\
       --network $NETWORK_NAME		\
       -p 8000:8000			\
       -e DATABASE_NAME=$DATABASE_NAME	\
       -e DATABASE_USER=$DATABASE_USER	\
       -e DATABASE_PASS=$DATABASE_PASS	\
       -e DATABASE_HOST=$DATABASE_HOST	\
       -e DATABASE_PORT=$DATABASE_PORT	\
       bookstore-saas:latest

podman run --rm -it			\
       --name web-init			\
       --network bookstore		\
       -e DATABASE_NAME=$DATABASE_NAME	\
       -e DATABASE_USER=$DATABASE_USER	\
       -e DATABASE_PASS=$DATABASE_PASS	\
       -e DATABASE_HOST=$DATABASE_HOST	\
       -e DATABASE_PORT=$DATABASE_PORT	\
       bookstore-saas:latest		\
       /bin/bash -c "until nc -zvw10 ${DATABASE_HOST:-postgres} ${DATABASE_PORT:-5432}; do echo 'waiting for database...'; sleep 2; done && python manage.py migrate"
