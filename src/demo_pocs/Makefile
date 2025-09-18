#

OBJ=postgres_testdb search_coarse_grain_api search_wardley search_coarse_grain search_coarse_grain2  search_coarse_grain_api dotest_entirecone dotest_getnodes definecontext

all: $(OBJ)

% : %.go ../../pkg/SSTorytime/SSTorytime.go
	go build -o $@ $@.go


clean:
	rm -f $(OBJ)
	rm -f *~

