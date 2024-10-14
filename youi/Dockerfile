# Base stage
FROM bitnami/minideb:latest AS base

RUN install_packages ca-certificates curl git

# Dynamically set the GOARCH and URL for Go binary download
ARG TARGETARCH

ENV PATH="/usr/local/go/bin:${PATH}"
ENV GOFLAGS=-buildvcs=false

# Development stage
FROM base AS dev

ENV DBUS_SESSION_BUS_ADDRESS autolaunch:

RUN install_packages wget gnupg \
	&& install_packages nodejs npm \
	&& npm install -g bun \
	&& install_packages libwebkit2gtk-4.0-dev build-essential curl wget libssl-dev libgtk-3-dev libayatana-appindicator3-dev librsvg2-dev

COPY . .

ENV PATH="/root/go/bin:${PATH}"

EXPOSE 1420

CMD ["bun", "run", "tauri", "dev"]

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
