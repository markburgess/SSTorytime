
# Debugging directly in postgres

Start with a shell `psql` and switch to the postgres user

<pre>
$ sudo su
$ su - postgres
$ psql sstoryline

\du  - list all databases
\l   - liste databases and types

\dt <type> show table type
\df show functions

select s from Node where s like '%fish%';

</pre>

