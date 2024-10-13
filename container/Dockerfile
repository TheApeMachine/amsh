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
    jq \
    unzip \
    zip

# Add the entrypoint script to handle user creation and custom message display
COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
CMD ["bash", "-l"]