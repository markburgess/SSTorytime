#!/bin/sh
#
# Very simplistic script for RAMify the postgres instance
# Should work on the standardized Linux Distros

echo Get root user credentials
sudo mkdir -p /mnt/pg_ram
echo Creating RAM disk
echo Mounting ram disk
if mount | grep -q /mnt/pg_ram; then 
   echo ram disk already mounted
else 
   sudo mount -t tmpfs -o size=800M tmpfs /mnt/pg_ram
fi

if /usr/bin/id postgres > /dev/null; then
 sudo chown postgres:postgres /mnt/pg_ram
 echo Stopping existing postgres
 sudo systemctl stop postgresql
 echo initialize database
 sudo su postgres -c "/usr/lib/postgresql17/bin/initdb -D /mnt/pg_ram/pgdata" > /dev/null
 echo restarting postgres
 sudo su postgres -c "/usr/lib/postgresql17/bin/pg_ctl -D /mnt/pg_ram/pgdata -l /mnt/pg_ram/logfile start"
else
 echo "Couldn't find a postgres user"
fi



