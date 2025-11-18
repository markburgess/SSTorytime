#!/bin/sh
#
# A script to configure

echo ""
echo Trying to configure the default database
echo ""

if sudo su postgres -c 'echo "\df" | psql sstoryline' > /dev/null; then
   echo Seems this may already have been created
   echo "\\df" | sudo su postgres -c "psql sstoryline"
else
   echo Trying to create the default sstoryline database
   echo "CREATE USER sstoryline PASSWORD 'sst_1234' superuser; CREATE DATABASE sstoryline; GRANT ALL PRIVILEGES ON DATABASE sstoryline TO sstoryline; CREATE EXTENSION UNACCENT;" | sudo su postgres -c "psql"
   echo "Done"
fi

echo \\df | sudo su postgres -c "psql sstoryline"

echo "* * * * * * * * * * * * * * * * * * * * * * * * * * *"
echo ""
echo "Now go to the examples directory and"
echo "make"
echo "to populate with some examples" 

