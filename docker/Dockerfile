# Minimal Linux image for the SSTorytime local test environment.
# Used by both the `workspace` and `web` services (see compose.yaml).
# Not a production image.

FROM golang:1.24.2-bookworm

RUN apt-get update \
 && apt-get install -y --no-install-recommends openssl jq ca-certificates curl postgresql-client \
 && rm -rf /var/lib/apt/lists/*
