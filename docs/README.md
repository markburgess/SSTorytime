# Setting up and getting started on a Linux computer

*At the moment, STTorytime is only supported on Linux based operating systems. Packaging for other systems will follow when things are more finished.*

## Summary

These are the things you will need to do:

* Download this repository, which contains examples of data input
languages N4L and examples of scripting your own programs.

* Install the `postgres` database, `postgres-contrib` extensions, and `psql` shell command line client.

* You need to make a decision about authentication credentials for the database. For personal use on a personal device, everything is local and private so there is no real need to set complex passwords for privacy. However, if you are setting up a shared resource, you might want to change the name of the database, user, and mickymouse password etc. That requires an extra step, changing the defaults and creating a file `$HOME/.SSTorytime` with those choices in your home directory.

* Install the Go(lang) programming and build environment.

* Get started by uploading ready-made examples.

* Read [Related series about semantic spacetime](https://mark-burgess-oslo-mb.medium.com/list/semantic-spacetime-and-data-analytics-28e9649c0ade)

*Note about troubleshooting: the "hard part" of setting up is to work around the quirks of the `Go` language and the database `Postgresql`. These are both delicate beasts: when they work they will just work, but if they don't they are very hard to debug. Postgres, in particular, fails silently and mysteriously. It keeps log files in `/var/lib/pgsql/data/log`. Luckily the major linux distros are mostly similar these days, so cross fingers that these instructions work. If you experience problems with the go language,
you may need to turn off modules:*

```
go env -w GO111MODULE=off
```



## Installing database Postgres

Hard part first; the postgres database is a bit of a monster. There are several steps to install it an set it up. Here's the summary:

* Use your local package manager to download and install packages for `postgres databaser server` and `psql client`.
* In postgres, you need root privileges to configure and create a database.
* Locate and edit the configuration file `pg_hba.conf` and make sure it's owned by the `postgres` user.
* Set the server to run in your systemd configuration.

You need privileged `root` access to access the postgres management account. Postgres prefers you to do everything as the postgres user not as root.

* To begin with, you need to start the database as root.
If this command doesn't work, check your local Linux instruction page as distros vary.

```
$ sudo systemctl enable postgresql
$ sudo systemctl start postgresql

$ ps waux | grep postgres
```
You should now see a number of processes running as the postgres user.

* * (as postgres user) Next login to the postgres user account and run the `psql` command there to gain root access:

```
sudo su -
su - postgres
psql
```

Only postgres user can CREATE or DROP a database.

* * (as postgres user) Set up a database for the examples. The default name in the code is:

```
\h for help

CREATE USER sstoryline PASSWORD 'sst_1234' superuser;
CREATE DATABASE sstoryline;
GRANT ALL PRIVILEGES ON DATABASE sstoryline TO sstoryline;
CREATE EXTENSION UNACCENT;
```

For the last line, you must have installed the extension packages `postgres-contrib`.
If you want to use psql to examine and manage
the database yourself using psql, it's useful to add your own account to the privileges, like this:

```
CREATE USER myusername;
GRANT ALL PRIVILEGES ON DATABASE sstoryline TO myusername;
\l
```

The `\l` command lists the databases, and you should now see the database.

* * (as postgres user) Locate the file `locate pg_hba.conf` for your distribution (you might have to search for it) and edit it as the postgres user.


```
$ myfavouriteeditor /var/lib/pgsql/data/pg_hba.conf

# TYPE  DATABASE        USER            ADDRESS                 METHOD

# "local" is for Unix domain socket connections only
local   all             all                                     peer
# IPv4 local connections:
host    all             all             127.0.0.1/32            <b>password</b>
# IPv6 local connections:
host    all             all             ::1/128                 <b>password</b>
```

This will allow you to connect to the database using the shell command `psql` command using password
authentication. Think of a suitable password.

Note that, if you accidentally edit the file as root, the owner of the file will be changed and postgres will fail to start.


Notice that the `psql` is a tool that accepts commands of two kind: backslash commands, e.g. describe tables for the current database `\dt`,  `\d tablename`, and describing stored functions `\df`. Also note that direct SQL commands, which must end in a semi-colon `;`.

* You should now be able to exit su log in to the postgres shell as an ordinary user, without sudo. Tap CTRL-D twice to get back to your user shell.
When connecting in code, you have to add the password. For a shell user, postgres recognizes your local
credentials.

```
$ psql sstoryline
```

*Cleary this is not a secure configuration, so you should only use this for testing on your laptop.
Also, note that this will not allow you to login until you also open up the configuration of postgres
as below.*

* IF YOU WANT TO CHANGE THE DATABASE CREDENTIALS from the defaults, by creating a file with these lines into a file `$HOME/.SSTorytime` :

```
dbname: my_sstoryline
user: my_sstoryline_user
passwd: new_password_for_sst_1234
```

Postgres is finnicky if you're not used to running it, but once these details are set up
you will be able to use the software. If you're planning to run a publicly available server, you
should learn more about the security of postgres. We won't go into that here.

## Installing the Go programming language for building and scripting

See also about Go

```
https://golang.org/dl/
```
After installing a package for your operating system, you need to set up some things in your environment so that you can forget about golang for the rest of your tortured life. One less thing to fret over.

You’ll need a command window (shell).
Then create some directories for the Golang workspace.
These are used to simplify the importing of packages. Finally, you need to link a gopath to your code download area.

```
% mkdir -p ~/go/bin
% mkdir -p ~/go/src
% git clone https://github.com/markburgess/SSTorytime
% ln -s ~/clonedirectory/pkg/SST ~/go/src/SST
```

The last step links the directory where you will keep the Smart Spacetime code library to the list of libraries that Go knows about. You’ll also need to set a GOPATH environment variable and add the installation directory to your execution path.For Linux (using default bash shell) you edit the file “~/.bashrc” in your home directory using your favourite text editor. It should contain these lines, as per the golang destructions:

```
export PATH=$PATH:/usr/local/go/bin
export GOPATH=~/go
```

Don’t forget to restart your shell or command window after editing this.

Since version 1.13 of Go, big changes have been made (and are expected to continue going forwards, sigh) concerning “modules” design. Unless you know what you’re doing, disable modules by running:

```
% go env -w GO111MODULE=off
```

To use the Go Driver, download it

```
% go get github.com/lib/pq

```

Try writing some simple programs in golang to learn its quirks. The
most annoying of these is the forced placement of curly braces and
indentations.

## Uploading the ready-made examples

Now that everything is working, simply do the following to try out the examples in the documentation:

```
$ cd examples
$ make
$ ../src/N4L-db -u LoopyLoo.n4l
```
