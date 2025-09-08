#

OBJ=text2N4L N4L searchN4L removeN4L http_server pathsolve notes graph_report API_EXAMPLE_1 API_EXAMPLE_2 API_EXAMPLE_3 API_EXAMPLE_4

all: $(OBJ)

removeN4L: removeN4L.go  ../pkg/SSTorytime/SSTorytime.go
	go build -o $@ $@.go

text2N4L: text2N4L.go  ../pkg/SSTorytime/SSTorytime.go
	go build -o $@ $@.go

N4L: N4L.go ../pkg/SSTorytime/SSTorytime.go
	go build -o $@ $@.go

searchN4L: searchN4L.go ../pkg/SSTorytime/SSTorytime.go
	go build -o $@ $@.go

http_server: http_server.go ../pkg/SSTorytime/SSTorytime.go
	go build -o $@ $@.go

notes: notes.go ../pkg/SSTorytime/SSTorytime.go
	go build -o $@ $@.go

pathsolve: pathsolve.go ../pkg/SSTorytime/SSTorytime.go
	go build -o $@ $@.go

graph_report: graph_report.go ../pkg/SSTorytime/SSTorytime.go
	go build -o $@ $@.go

API_EXAMPLE_1: API_EXAMPLE_1.go ../pkg/SSTorytime/SSTorytime.go
	go build -o $@ $@.go

API_EXAMPLE_2: API_EXAMPLE_2.go ../pkg/SSTorytime/SSTorytime.go
	go build -o $@ $@.go

API_EXAMPLE_3: API_EXAMPLE_3.go ../pkg/SSTorytime/SSTorytime.go
	go build -o $@ $@.go

API_EXAMPLE_4: API_EXAMPLE_4.go ../pkg/SSTorytime/SSTorytime.go
	go build -o $@ $@.go

clean:
	rm -f $(OBJ)
	rm -f *~ demo_pocs/*~

% : %.go
	go build $<
