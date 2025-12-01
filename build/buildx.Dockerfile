# syntax=docker/dockerfile:1.4

# Build stage: compile the binary for the target platform
FROM --platform=$BUILDPLATFORM golang:1.25.4 AS builder

ENV GOROOT=/usr/local/go \
	GOTOOLCHAIN=auto \
	CGO_ENABLED=0

WORKDIR /src

# Cache modules separately
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
	--mount=type=cache,target=/root/.cache/go-build \
	go mod download

# Copy the rest of the source
COPY . .

# Build for the requested target OS/ARCH
ARG TARGETOS
ARG TARGETARCH
RUN --mount=type=cache,target=/go/pkg/mod \
	--mount=type=cache,target=/root/.cache/go-build \
	GOOS="$TARGETOS" GOARCH="$TARGETARCH" go build -o /out/edubase-to-pdf ./

# Runtime stage: include system deps and the compiled binary
FROM --platform=$TARGETPLATFORM debian:bookworm-slim

# Metadata as defined in OCI image spec annotations
LABEL org.opencontainers.image.vendor="michaelbeutler"
LABEL org.opencontainers.image.title="edubase-to-pdf"

# Install Playwright browser runtime dependencies
RUN set -eux; \
		apt-get update; \
		apt-get install -y --no-install-recommends \
			ca-certificates \
			fonts-liberation \
			libasound2 \
			libatk-bridge2.0-0 \
			libatk1.0-0 \
			libatspi2.0-0 \
			libcairo2 \
			libcups2 \
			libdbus-1-3 \
			libdrm2 \
			libgbm1 \
			libglib2.0-0 \
			libnspr4 \
			libnss3 \
			libpango-1.0-0 \
			libx11-6 \
			libxcomposite1 \
			libxdamage1 \
			libxext6 \
			libxfixes3 \
			libxkbcommon0 \
			libxrandr2 \
			libxcb1; \
	rm -rf /var/lib/apt/lists/*

COPY --from=builder /out/edubase-to-pdf /usr/bin/edubase-to-pdf

CMD ["edubase-to-pdf"]