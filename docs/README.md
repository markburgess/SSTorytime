
# Setting up and getting started

To do:

* Download this repository, which contains examples of data input
languages N4L and examples of scripting your own programs.

* Install the `postgres` database, `postgres-contrib` extensions, and `psql` shell command line client.

* Install the Go(lang) programming and build environment
<pre>
    cd pkg/SSTorytime ; go mod init SSTorytime
	cd src ; go mod init main ;
	cd src ; go mod edit -replace SSTorytime=../pkg/SSTorytime
</pre>

* [Related series about semantic spacetime](https://mark-burgess-oslo-mb.medium.com/list/semantic-spacetime-and-data-analytics-28e9649c0ade)

## Running on a GNU/Linux distribution

Once you've installed the dependencies: Go programming language, Postgres database, the Postgress-contrib library, and (optionally) the Make program, you could start like this:
<pre>
$ make
$ cd examples
$ ../src/N4L-db -u chinese*n4l Mary.n4l doors.n4l doubleslit.n4l brains.n4l

2442:N4L chinese.n4l WARNING: Found a note to self in the text (ZZZZZZZZ) at line 2442 

2517:N4L chinese.n4l WARNING: Found a note to self in the text (HERE TO DO) at line 2517 
Uploading nodes..
Storing nodes...
.................................................................................................................................................................................................................................................................................................................................................................................................................................................................................................................................................................................................................................................................................................................

</pre>
Once the data are uploaded, you can try a simple test, e.g. the demo program
<pre>
% cd src/demo_pocs
% go run search_noninteractive.go
--------------------------------------------------
Looking for relevant arrows by [poem] chinese
--------------------------------------------------
    (No relevant matches)
--------------------------------------------------
Looking for relevant nodes by tiger
--------------------------------------------------
Looking for nodes like tiger...   Found possible relevant nodes: [{4 626}]

    -------------------------------------------
     Search text MATCH #1 via -LeadsTo connection
     (search tiger => hit two tigers, two tigers)
    -------------------------------------------
     - SSType -LeadsTo  cone item:  two tigers, two tigers , found in notes on chinese
     - SSType -LeadsTo  cone item:  两只老虎, 两只老虎 , found in notes on chinese
     - SSType -LeadsTo  cone item:  Liǎng zhī lǎohǔ, liǎng zhī lǎohǔ , found in notes on chinese
     (END 1)

--------------------------------------------------
checking whether any arrows also match search tiger (in any context)
--------------------------------------------------
mark% 

</pre>
If you see something like this, everything is working.

### Installing database Postgres

* Use your local package manager to download packages for `postgres databaser server` and `psql client`.
* In postgres, you need root privileges to configure and create a database.
* Set the server to run in your systemd configuration.

<pre>
sudo su -
su - postgres
</pre>
Once you have a root shell, you can grant access to postgres to other users.

* You will normally access postgres for a specific database that you create once and for all, as root user.

* `psql` is a tool that accepts commands of two kind:

 * Backslash commands, e.g. describe tables for the current database `\dt`,  `\d tablename`, and describing stored functions `\df`.
 * As direct SQL commands, which must end in ;

Set up a database for the examples, e.g. as root user. The default name in the code is:
<pre>
\h for help

CREATE user sstoryline password 'sst_1234' superuser;
CREATE DATABASE sstoryline;
CREATE DATABASE newdb;
GRANT ALL PRIVILEGES ON DATABASE sstoryline TO sstoryline;
GRANT ALL PRIVILEGES ON DATABASE newdb TO sstoryline;
CREATE EXTENSION UNACCENT;
</pre>
For the last line, you must have installed the extension packages `postgres-contrib`.

* In the examples, two databases are used: `sstoryline` and `newdb` for personal scripting and testing,
it's useful to have another, called `newdb`.
* Only superuser can CREATE or DROP a database.

* You should now be able to log in to the postgres shell as an ordinary user, without sudo.

<pre>
psql newdb
psql sstoryline
</pre>
When connecting in code, you have to add the password. For a shell user, postgres recognizes your local
credentials.

Cleary this is not a secure configuration, so you should only use this for testing on your laptop.
Also, note that this will not allow you to login until you also open up the configuration of postgres
as below. In summary, 

* * Create a database.
* * Create a user for accessing the database over a local (or later remote) socket.
* * Grant access and permissions to the SST user, by editing a configuration file.
* * Locate the file `locate pg_hba.conf` for your distribution (you might have to search for it):

<pre>

# TYPE  DATABASE        USER            ADDRESS                 METHOD

# "local" is for Unix domain socket connections only
local   all             all                                     peer
# IPv4 local connections:
host    all             all             127.0.0.1/32            <b>password</b>
# IPv6 local connections:
host    all             all             ::1/128                 <b>password</b>
</pre>
This will allow you to connect to the database using the shell command `psql` command using password
authentication. Think of a suitable password.



Postgres is finnicky if you're not used to running it, but once these details are set up
you will be able to use the software. If you're planning to run a publicly available server, you
should learn more about the security of postgres. We won't go into that here.



### Installing the Go programming language for building and scripting

See also about Go
<pre>
https://golang.org/dl/
</pre>
After installing a package for your operating system, you need to set up some things in your environment so that you can forget about golang for the rest of your tortured life. One less thing to fret over.

You’ll need a command window (shell). 
Then create some directories for the Golang workspace. 
These are used to simplify the importing of packages. Finally, you need to link a gopath to your code download area.
<pre>
% mkdir -p ~/go/bin
% mkdir -p ~/go/src
% git clone https://github.com/markburgess/SemanticSpaceTime
% ln -s ~/clonedirectory/pkg/SST ~/go/src/SST
</pre>
The last step links the directory where you will keep the Smart Spacetime code library to the list of libraries that Go knows about. You’ll also need to set a GOPATH environment variable and add the installation directory to your execution path.For Linux (using default bash shell) you edit the file “~/.bashrc” in your home directory using your favourite text editor. It should contain these lines, as per the golang destructions:
<pre>
export PATH=$PATH:/usr/local/go/binexport GOPATH=~/go# Set a short promptexport PS1=”mark% “
</pre>
Don’t forget to restart your shell or command window after editing this.

Since version 1.13 of Go, big changes have been made (and are expected to continue going forwards, sigh) concerning “modules” design. Unless you know what you’re doing, disable modules by running:
<pre>
% go env -w GO111MODULE=off
</pre>
To use the Go Driver, download it
<pre>
% go get github.com/lib/pq

</pre>

Try writing some simple programs in golang to learn its quirks. The
most annoying of these is the forced placement of curly braces and
indentations.



