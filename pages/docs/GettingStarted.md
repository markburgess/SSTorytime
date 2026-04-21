# Setting up and getting started

![A retro CRT terminal on a wooden desk flanked by a PostgreSQL elephant and a Go gopher, connected by ochre flow arrows.](figs/getting_started_hero.jpg){ align=center }

As we figure out how to make the project easily available, you will need to compile
the programs by following these destructions below. You can get started with or without Docker (for those who know). You may need to install these prerequisites:

* Git
* Make
* Postgres
* Docker (optional, e.g. Docker desktop)

First compile the software. From the project root, run:

<pre>
$ make
</pre>

The build drops every binary into `src/bin/`. Put that directory on your `$PATH`
so the tools below — `N4L`, `searchN4L`, `http_server`, `pathsolve`, `notes`,
`graph_report`, `text2N4L`, `removeN4L` — work as bare commands:

<pre>
$ export PATH=$PWD/src/bin:$PATH    # so `N4L`, `searchN4L`, `http_server` work directly
</pre>

The rest of this page, the [Tutorial](Tutorial.md), and the cookbooks all
assume you've done this. If you prefer not to, substitute `src/bin/<tool>`
(from the repo root) wherever you see a bare tool name.

Next, setup the PostgreSQL database:
<pre>
$ make db
</pre>
or (the preferred option for Linux):
<pre>
$ make ramdb
</pre>
Now that you've compiled and the database is running, you can upload the example data to play with:
<pre>
$ cd examples/
$ make
</pre>
With data, you can now run the web server:
<pre>
$ http_server
</pre>
and open a web browser at `https://localhost:8443`. (The plaintext port `:8080` will issue a 301 redirect to the HTTPS URL.) You'll see a self-signed-certificate warning the first time — accept it for local development. Try searching for SSTorytime!

## 1. Find your operating system

Here is the rough plan:

* **GNU/Linux**: Use your local package manager to install `postgres,git,make`, then download Go(lang), then compile the software

* **MacOS**: Use `homebrew` to install `go,make,git`, then download Docker Desktop, use docker to run postgres, compile software

* **Windows**: Use docker to run postgres, download git and go, compile software.


## 2. Install Postgres database

### 2a. Steps for running postgres in a Docker container

The PostgreSQL database dependency can by run in a Docker container to avoid local installation and configuration. See [Running the SSTorytime database in docker](https://github.com/markburgess/SSTorytime/blob/main/postgres-docker/README.md) for further details.

### or 2b. Steps for postgres with package installation

The postgres database is a bit of a monster. There are several steps to install it an set it up. 
You can choose to keep the database in RAM memory or on disk. Running in RAM is recommended unless you are 
already using it for something else, since SSTorytime uses postgres basically as a cache and some SSDs wear
out more quickly with heavy usage from a read/write database.

Here's the summary:

* Use your local package manager to download and install packages for `postgres databaser server`, `postgres-contrib` extensions, and `psql` shell command line client.

* 1. To configure postgres, you need root (sudo) privileges to configure and create a database before you can being.
* 2. Locate and edit the configuration file `pg_hba.conf` and make sure it's owned by the `postgres` user.
* 3. Set the server to run in your systemd configuration.

You need privileged `root` access to access the postgres management account. Postgres prefers you to do everything as the postgres user not as root.


## 3. Setting up the SST database in postgres - two methods

You can set up postgres directly or run it in RAM disk memory. Running
in a RAM disk is fast and protects your storage device (SSD or
harddisk) from unnecessary wear while reloading and changing data a
lot.  If you choose a RAM disk, rebooting the computer or powering off
will lose all the data in the database.  However, if you are only
using the database to keep N4L notes, you can rebuild it anytime from
source.


### 3a. SST Postgres in RAM disk memory [Linux]

You can install Postgres in memory to increase performance of the
upload and search, and to preserve your laptop SSD disks. The downside
is that each time you reboot you will have to repeat this procedure
and all will be lost.

- To do so, create a new data folder, and mount it as a memory file system.
- grant access rights to your postgres user.
- stop the default postgres system service.
- start manually postgres using your new filesystem as data storage, or configure the postgres service to use the new memory data folder

**Beware !**: all data in the postgres database will be lost when restarting processes. 
But you can always rebuild the schema, and reload your data graph from your N4L files using the tool N4L.
e.g. paste in the following commands to a shell, giving the root password:

!!! warning "PostgreSQL 14+ required; system-managed install assumed"
    SSTorytime needs PostgreSQL **14 or newer**. The snippet below resolves the
    Postgres binary directory dynamically (via `pg_config` or the location of
    `initdb` on `$PATH`), which works across Fedora 43 (pg 16), Debian/Ubuntu
    22.04+ (pg 14+), and macOS Homebrew. It assumes you installed Postgres via
    your system package manager; if you are running Postgres inside Docker,
    use the RAM-disk recipe in the container instead.

```
sudo su -

mkdir -p /mnt/pg_ram
mount -t tmpfs -o size=800M tmpfs /mnt/pg_ram
chown postgres:postgres /mnt/pg_ram
systemctl stop postgresql
su postgres -

# Resolve the Postgres bin directory for whichever major version is installed.
PG_BINDIR=$(pg_config --bindir 2>/dev/null || dirname $(which initdb 2>/dev/null))

$PG_BINDIR/initdb -D /mnt/pg_ram/pgdata
$PG_BINDIR/pg_ctl -D /mnt/pg_ram/pgdata -l /mnt/pg_ram/logfile start

```

Now repeat the setup steps for the database:

```
$ sudo su -  
(root password)
# su - postgres
## psql
CREATE USER sstoryline PASSWORD 'sst_1234' superuser;
CREATE DATABASE sstoryline;
GRANT ALL PRIVILEGES ON DATABASE sstoryline TO sstoryline;
CREATE EXTENSION UNACCENT;
```


* In the long run, if running publicly, you will need to make a decision about authentication credentials for the database. For tesing, for personal use on a personal device, everything is local and private so there is no real need to set complex passwords for privacy. However, if you are setting up a shared resource, you might want to change the name of the database, user, and mickymouse password etc. That requires an extra step, changing the defaults and creating a file `$HOME/.SSTorytime` with those choices in your home directory.


### 3b. SST Postgres on secondary disk storage

To complete the setup, login to the postgres user account and run the `psql` command.
Only postgres user can CREATE or DROP a database. Since you probably don't know the postgres password,
you can go via the root account:

```
$ sudo su -  
(root password)
# su - postgres
## psql
```

You now have a postgres shell.
To set up a database,, simply paste in these commands:

```
CREATE USER sstoryline PASSWORD 'sst_1234' superuser;
CREATE DATABASE sstoryline;
GRANT ALL PRIVILEGES ON DATABASE sstoryline TO sstoryline;
CREATE EXTENSION UNACCENT;
```

For the last line, you must have installed the extension packages `postgres-contrib`.

The `\l` command lists the databases, and you should now see the database.


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


* In the long run, if running publicly, you will need to make a decision about authentication credentials for the database. For tesing, for personal use on a personal device, everything is local and private so there is no real need to set complex passwords for privacy. However, if you are setting up a shared resource, you might want to change the name of the database, user, and mickymouse password etc. That requires an extra step, changing the defaults and creating a file `$HOME/.SSTorytime` with those choices in your home directory.

### Configuring Postgres for client internet access

To begin with, you need to start the database as root.
If this command doesn't work, check your local Linux instruction page as distros vary.

```
$ sudo systemctl enable postgresql
$ sudo systemctl start postgresql

$ ps waux | grep postgres
```
You should now see a number of processes running as the postgres user.

* To complete the setup you need to locate the file `locate pg_hba.conf` for your distribution (you might have to search for it) and edit it as the postgres user and edit it go grant connection access.

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
authentication. 

Note that, if you accidentally edit the file as root, the owner of the file will be changed and postgres will fail to start.


Notice that the `psql` is a tool that accepts commands of two kind: backslash commands, e.g. describe tables for the current database `\dt`,  `\d tablename`, and describing stored functions `\df`. Also note that direct SQL commands, which must end in a semi-colon `;`.



## 4. Installing the Go programming language for building and scripting

The Go language is easy like "Python" but fast and strongly typed, with compiler
checks. You can think of it as a "better Python" — in spite of some questionable
aesthetic choices. Grab an installer for your operating system from
[golang.org/dl](https://golang.org/dl/). Any Go 1.24 or newer is fine.

After installing, make sure `go version` works in a fresh shell. That's all the
setup you need — the project builds straight out of the source checkout via
`make`, which uses the standard `go.mod` in the repo root. There is **no
GOPATH symlink** required and no need to touch `~/go/src`.

## Uploading the ready-made examples

Now that everything is working, simply do the following to try out the examples in the documentation:

```
$ cd examples
$ make
$ N4L -u LoopyLoo.n4l
```
