# SSTorytime — standard Go layout
# Primary binary: bin/sstorytime (Cobra + busybox multicall).
# Packagers / Nix: `make build` or `go build -o bin/sstorytime ./cmd/sstorytime`

BINDIR ?= bin
GO     ?= go

# Single multicall binary
SSTORYTIME = $(BINDIR)/sstorytime

# Historical tool names (symlinks → sstorytime)
MULTICALL_LINKS = \
	N4L \
	searchN4L \
	removeN4L \
	text2N4L \
	notes \
	pathsolve \
	graph_report \
	http_server \
	API_EXAMPLE_1 \
	API_EXAMPLE_2 \
	API_EXAMPLE_3 \
	API_EXAMPLE_4 \
	definecontext \
	dotest_entirecone \
	dotest_getnodes \
	postgres_testdb

.PHONY: all build tools demos test clean css db ramdb sstorytime links

all: build

build: sstorytime links

sstorytime: $(SSTORYTIME)

$(SSTORYTIME):
	@mkdir -p $(BINDIR)
	$(GO) build -o $@ ./cmd/sstorytime

# Symlinks so argv0 multicall works (N4L → sstorytime, …)
links: $(SSTORYTIME)
	@for name in $(MULTICALL_LINKS); do \
		ln -sfn sstorytime $(BINDIR)/$$name; \
	done

# Backward-compatible aliases
tools: build
demos: build

css:
	cd internal/app/httpserver && $(GO) run -tags css_tool ./css-builder

test: $(SSTORYTIME)
	$(GO) test ./...
	$(MAKE) -C tests test

clean:
	rm -rf $(BINDIR)
	-$(MAKE) -C tests clean
	-$(MAKE) -C examples clean

db:
	sh contrib/makedb.sh

ramdb:
	(cd contrib; sh ramify.sh)
	(cd contrib; sh makeramdb.sh)
