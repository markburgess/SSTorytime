#!/bin/sh
#
# A script to configure

if /usr/bin/id postgres > /dev/null; then
 sudo chown postgres:postgres /mnt/pg_ram
 echo Stopping existing postgres
 sudo systemctl stop postgresql
 echo ""
 echo initialize database

 if [ -f /mnt/pg_ram/pgdata ]; then
    echo The database /mnt/pg_ram/pgdata already exists, remove and try again?
    exit 1
 else
    if sudo su postgres -c "/usr/lib/postgresql17/bin/initdb -D /mnt/pg_ram/pgdata" > /dev/null; then
       echo done
    else
       echo ""
       echo "Looks like postgres is already running"
    fi
 fi

 echo restarting postgres
 sudo su postgres -c "/usr/lib/postgresql17/bin/pg_ctl -D /mnt/pg_ram/pgdata stop"
 sudo su postgres -c "/usr/lib/postgresql17/bin/pg_ctl -D /mnt/pg_ram/pgdata -l /mnt/pg_ram/logfile start"
 echo ""
 echo -n "Ramdisk: "
 mount | grep /mnt/pg_ram
else
 echo "Couldn't find a postgres user"
 exit 1
fi

echo ""
echo Now try to configure the default database
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

