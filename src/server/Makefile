all:
	mkdir -p ../bin
	go run css-builder/css_template.go
	./make_certificate
	go build -o ../bin/http_server ./...
