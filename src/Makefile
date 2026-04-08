#

OBJ=bin/text2N4L bin/N4L bin/searchN4L bin/removeN4L bin/http_server bin/pathsolve bin/notes bin/graph_report bin/API_EXAMPLE_1 bin/API_EXAMPLE_2 bin/API_EXAMPLE_3 bin/API_EXAMPLE_4 demo_pocs/bin/postgres_testdb demo_pocs/bin/dotest_getnodes demo_pocs/bin/dotest_entirecone demo_pocs/bin/definecontext

all: $(OBJ)

demo_pocs/bin/dotest_getnodes: demo_pocs/dotest_getnodes/dotest_getnodes.go ../pkg/SSTorytime
	cd demo_pocs/dotest_getnodes ; make

demo_pocs/bin/dotest_entirecone: demo_pocs/dotest_entirecone/dotest_entirecone.go ../pkg/SSTorytime
	cd demo_pocs/dotest_entirecone ; make

demo_pocs/bin/definecontext: demo_pocs/definecontext/definecontext.go ../pkg/SSTorytime
	cd demo_pocs/definecontext ; make

demo_pocs/bin/postgres_testdb: demo_pocs/postgres_testdb/postgres_testdb.go ../pkg/SSTorytime
	cd demo_pocs/postgres_testdb ; make

bin/removeN4L: removeN4L/removeN4L.go ../pkg/SSTorytime
	cd removeN4L ; make

bin/text2N4L: text2N4L/text2N4L.go ../pkg/SSTorytime
	cd text2N4L ; make

bin/N4L: N4L/N4L.go ../pkg/SSTorytime
	cd N4L ; make

bin/searchN4L: searchN4L/searchN4L.go ../pkg/SSTorytime
	cd searchN4L ; make

bin/http_server: server/http_server.go ../pkg/SSTorytime server/public
	cd server ; make

bin/notes: notes/notes.go ../pkg/SSTorytime
	cd notes ; make

bin/pathsolve: pathsolve/pathsolve.go ../pkg/SSTorytime
	cd pathsolve ; make

bin/graph_report: graph_report/graph_report.go ../pkg/SSTorytime
	cd graph_report ; make

bin/API_EXAMPLE_1: API_EXAMPLE_1/API_EXAMPLE_1.go ../pkg/SSTorytime
	cd API_EXAMPLE_1 ; make

bin/API_EXAMPLE_2: API_EXAMPLE_2/API_EXAMPLE_2.go ../pkg/SSTorytime
	cd API_EXAMPLE_2 ; make

bin/API_EXAMPLE_3: API_EXAMPLE_3/API_EXAMPLE_3.go ../pkg/SSTorytime
	cd API_EXAMPLE_3 ; make

bin/API_EXAMPLE_4: API_EXAMPLE_4/API_EXAMPLE_4.go ../pkg/SSTorytime
	cd API_EXAMPLE_4 ; make

clean:
	rm -f $(OBJ)
	rm -f *~ \#* demo_pocs/*~ demo_pocs/\#~
