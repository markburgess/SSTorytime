BINARY := ../src/N4L

all: $(BINARY)
	$(BINARY) -u -wipe doors.n4l Mary.n4l chinese*n4l branches.n4l doubleslit.n4l ConstructionProcesses.n4l wardleymap.n4l brains.n4l kubernetes.n4l SSTorytime.n4l SSTsearchexamples.n4l integral.n4l reasoning.n4l moon.n4l astronomy.n4l promises.n4l json_example_1.n4l json_example_2.n4l NDA.n4l reminders.n4l # Darwin.n4l MobyDickNotes.n4l

$(BINARY):
	@echo "Binary $(BINARY) not found. Building from project root ..."
	$(MAKE) -C ../.

clean:
	rm -f *_test_log *~

