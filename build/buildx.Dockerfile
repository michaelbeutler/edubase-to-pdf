# syntax=docker/dockerfile:1.4
FROM --platform=$TARGETPLATFORM golang:1.25.4

# Metadata as defined in OCI image spec annotations
LABEL org.opencontainers.image.vendor="michaelbeutler"
LABEL org.opencontainers.image.title="edubase-to-pdf"

ENV GOROOT /usr/local/go

# Allow to download a more recent version of Go.
# https://go.dev/doc/toolchain
# GOTOOLCHAIN=auto is shorthand for GOTOOLCHAIN=local+auto
ENV GOTOOLCHAIN auto

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

COPY edubase-to-pdf /usr/bin/
CMD ["edubase-to-pdf"]