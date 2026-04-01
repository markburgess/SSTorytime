#

OBJ=bin/text2N4L bin/N4L bin/searchN4L bin/removeN4L bin/http_server bin/pathsolve bin/notes bin/graph_report bin/API_EXAMPLE_1 bin/API_EXAMPLE_2 bin/API_EXAMPLE_3 bin/API_EXAMPLE_4 demo_pocs/bin/postgres_testdb demo_pocs/bin/dotest_getnodes demo_pocs/bin/dotest_entirecone demo_pocs/bin/define_context

all: $(OBJ)

demo_pocs/bin/dotest_getnodes:
	cd demo_pocs/dotest_getnodes ; make

demo_pocs/bin/dotest_entirecone:
	cd demo_pocs/dotest_entirecone ; make

demo_pocs/bin/define_context:
	cd demo_pocs/definecontext ; make

demo_pocs/bin/postgres_testdb:
	cd demo_pocs/postgres_testdb ; make

bin/removeN4L:
	cd removeN4L ; make

bin/text2N4L:
	cd text2N4L ; make

bin/N4L:
	cd N4L ; make

bin/searchN4L:
	cd searchN4L ; make

bin/http_server:
#	(cd server/css-builder; go run css_template.go)
	(cd server; ./make_certificate)
	cd server ; make

bin/notes:
	cd notes ; make

bin/pathsolve:
	cd pathsolve ; make

bin/graph_report:
	cd graph_report ; make

bin/API_EXAMPLE_1:
	cd API_EXAMPLE_1 ; make

bin/API_EXAMPLE_2:
	cd API_EXAMPLE_2 ; make

bin/API_EXAMPLE_3:
	cd API_EXAMPLE_3 ; make

bin/API_EXAMPLE_4:
	cd API_EXAMPLE_4 ; make

clean:
	rm -f $(OBJ)
	rm -f *~ \#* demo_pocs/*~ demo_pocs/\#~
