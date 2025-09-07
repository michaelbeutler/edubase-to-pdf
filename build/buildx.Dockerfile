# syntax=docker/dockerfile:1.4
FROM golang:1.25.1

# Metadata as defined in OCI image spec annotations
LABEL org.opencontainers.image.vendor="michaelbeutler"
LABEL org.opencontainers.image.title="edubase-to-pdf"

ENV GOROOT /usr/local/go

# Allow to download a more recent version of Go.
# https://go.dev/doc/toolchain
# GOTOOLCHAIN=auto is shorthand for GOTOOLCHAIN=local+auto
ENV GOTOOLCHAIN auto

COPY edubase-to-pdf /usr/bin/
CMD ["edubase-to-pdf"]