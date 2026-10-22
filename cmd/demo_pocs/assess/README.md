
# Scanning an agent run

Assuming a Linux environment, after downloading the Go environment, set up the respository:

```
$ git clone https://github.com/markburgess/SSTorytime.git
$ cd SSTorytime
$ make ramdb # prompts for (root passwd)
$ cd cmd/demo_poc/assess
```

## Agent Data

Collect the markdown definitions and json output traces of an agent into a directory and run:

```
$ go run assess.go -dir /path/to/mydirectory
```

Then upload the files

```
$ ../../bin/N4L -wipe -u AGENT_ASSESSMENT_GENERATED.n4l

```
Run the server:
```
$ cd SSTorytime/cmd/server
$ ../bin/http_server 

```
Now connect to `localhost:8080` and click `TOC` for the table of contents.