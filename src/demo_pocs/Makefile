
#

OBJ= bin/postgres_testdb bin/dotest_getnodes bin/dotest_entirecone bin/define_context

all: $(OBJ)

bin/dotest_getnodes:
	cd dotest_getnodes ; make

bin/dotest_entirecone:
	cd dotest_entirecone ; make

bin/define_context:
	cd definecontext ; make

bin/postgres_testdb:
	cd postgres_testdb ; make

clean:
	rm -f $(OBJ)
	rm -f *~ \#* \#~
