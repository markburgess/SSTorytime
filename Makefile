# SSTorytime — standard Go layout (cmd/, internal/)
# Packagers / Nix: `make build` or `go build -o bin/N4L ./cmd/N4L`

BINDIR ?= bin
GO     ?= go

TOOLS = \
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
	API_EXAMPLE_4

DEMOS = \
	definecontext \
	dotest_entirecone \
	dotest_getnodes \
	postgres_testdb

.PHONY: all build tools demos test clean css db ramdb

all: build

build: tools demos

tools: $(addprefix $(BINDIR)/,$(TOOLS))

demos: $(addprefix $(BINDIR)/,$(DEMOS))

$(BINDIR)/%:
	@mkdir -p $(BINDIR)
	$(GO) build -o $@ ./cmd/$*

css:
	cd cmd/http_server && $(GO) run -tags css_tool ./css-builder

test: tools
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
