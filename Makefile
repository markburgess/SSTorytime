
all: src/N4L test

src/bin/N4L:
	(cd src; make)
	(cd src/demo_pocs; make)

test: src/bin/N4L
	(cd src; make)
	(cd src/demo_pocs; make)
	(cd tests; make)
clean:
	rm -f *~ \#* N4L
	(cd src; make clean)
	(cd examples; make clean)
	(cd src/demo_pocs; make clean)

ramdisk:
ramdb:
	(cd contrib; sh ramify.sh)
	(cd contrib; sh makeramdb.sh)

db:
	sh contrib/makedb.sh
