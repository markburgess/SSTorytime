#!/bin/sh
#
# A script to configure

echo "CREATE USER sstoryline PASSWORD 'sst_1234' superuser; CREATE DATABASE sstoryline; GRANT ALL PRIVILEGES ON DATABASE sstoryline TO sstoryline; CREATE EXTENSION UNACCENT;" | sudo su postgres -c psql
