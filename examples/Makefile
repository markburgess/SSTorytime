# Prefer: go build -o ../bin/sstorytime ./cmd/sstorytime
BINARY := ../bin/sstorytime

all:
	@test -x $(BINARY) || (echo "build first: go build -o bin/sstorytime ./cmd/sstorytime" && exit 1)
	$(BINARY) examples load all

clean:
	rm -f *_test_log *~
