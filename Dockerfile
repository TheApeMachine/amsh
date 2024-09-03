# Base stage
FROM bitnami/minideb:latest AS base

RUN install_packages ca-certificates curl git

# Dynamically set the GOARCH and URL for Go binary download
ARG TARGETARCH
RUN if [ "${TARGETARCH}" = "arm64" ]; then \
	ARCH=arm64; \
	else \
	ARCH=amd64; \
	fi && \
	curl -LO https://golang.org/dl/go1.23.0.linux-${ARCH}.tar.gz && \
	tar -C /usr/local -xzf go1.23.0.linux-${ARCH}.tar.gz && \
	rm go1.22.0.linux-${ARCH}.tar.gz

ENV PATH="/usr/local/go/bin:${PATH}"
ENV GOFLAGS=-buildvcs=false

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod tidy && go mod download

# Development stage
FROM base AS dev
RUN go install github.com/air-verse/air@latest
COPY . .
ENV PATH="/root/go/bin:${PATH}"
EXPOSE 8080
CMD ["air"]

# Production stage
FROM base AS builder
COPY . .
RUN go build -o app

FROM bitnami/minideb:latest AS production
ARG STAGE=playground
ENV STAGE=$STAGE
WORKDIR /app
COPY --from=builder /app/app .
RUN install_packages ca-certificates \
	&& useradd -m -s /bin/bash nonroot \
	&& chown -R nonroot:nonroot /app
USER nonroot
EXPOSE 8080
CMD ./app serve --stage=${STAGE}