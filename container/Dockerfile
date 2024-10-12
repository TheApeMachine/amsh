# Base stage
FROM bitnami/minideb:latest AS base

# Install a comprehensive set of development tools and languages.
# This allows for a versatile development environment capable of
# handling a wide range of programming tasks.
RUN install_packages \
    sudo \
    ca-certificates \
    wget \
    curl \
    git \
    tmux \
    jq \
    unzip \
    zip \
    python3 \
    python3-pip \
    nodejs \
    npm \
    rustc \
    rustup \
    dotnet-sdk-7.0 \
    php \
    php-cli \
    php-curl \
    php-json \
    php-mbstring \
    php-xml \
    php-zip

# Dynamically install Go based on the target architecture.
# This ensures compatibility across different hardware platforms.
ARG TARGETARCH
RUN if [ "${TARGETARCH}" = "arm64" ]; then \
    ARCH=arm64; \
    else \
    ARCH=amd64; \
    fi && \
    curl -LO https://golang.org/dl/go1.22.0.linux-${ARCH}.tar.gz && \
    tar -C /usr/local -xzf go1.22.0.linux-${ARCH}.tar.gz && \
    rm go1.22.0.linux-${ARCH}.tar.gz

ENV PATH="/usr/local/go/bin:${PATH}"
ENV GOFLAGS=-buildvcs=false

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod tidy && go mod download

# Development stage
FROM base AS dev
RUN install_packages wget curl git tmux \
    && go install github.com/air-verse/air@latest
COPY . .
ENV PATH="/root/go/bin:${PATH}"
EXPOSE 8080

# Add the entrypoint script to handle user creation and custom message display
COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
CMD ["bash", "-l"]