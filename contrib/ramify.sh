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
   sudo mount -t tmpfs -o size=1000M tmpfs /mnt/pg_ram
fi

