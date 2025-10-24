# Multi-stage build
FROM golang:alpine AS builder

WORKDIR /build

# Copy everything needed for build
COPY go.mod ./
COPY pkg/ ./pkg/
COPY src/ ./src/

# Build all binaries from src directory
WORKDIR /build/src
RUN go build -o N4L N4L.go && \
    go build -o text2N4L text2N4L.go && \
    go build -o searchN4L searchN4L.go && \
    go build -o removeN4L removeN4L.go && \
    go build -o pathsolve pathsolve.go && \
    go build -o notes notes.go && \
    go build -o graph_report graph_report.go && \
    go build -o http_server ./server/http_server.go

# Runtime image
FROM alpine:latest

RUN apk --no-cache add ca-certificates postgresql-client make

WORKDIR /app

# Copy binaries from builder
COPY --from=builder /build/src/N4L ./
COPY --from=builder /build/src/text2N4L ./
COPY --from=builder /build/src/searchN4L ./
COPY --from=builder /build/src/removeN4L ./
COPY --from=builder /build/src/pathsolve ./
COPY --from=builder /build/src/notes ./
COPY --from=builder /build/src/graph_report ./
COPY --from=builder /build/src/http_server ./

# Copy config and examples
COPY SSTconfig/ ./SSTconfig/
COPY examples/ ./examples/

# Default to http_server
CMD ["./http_server"]
