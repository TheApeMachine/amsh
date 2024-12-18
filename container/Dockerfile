# Base stage
FROM bitnami/minideb:latest AS base

# Install a comprehensive set of development tools and languages.
# This allows for a versatile development environment capable of
# handling a wide range of programming tasks.
RUN install_packages \
    ca-certificates \
    wget \
    curl \
    git \
    jq \
    unzip \
    zip


WORKDIR /root/.ssh
RUN echo "Host github.com\n\tStrictHostKeyChecking no\n" >> /root/.ssh/config \
    && echo "Use /tmp/workspace to examine the code and try out fixes, then clone the actual repository (github.com/TheApeMachine/amsh) into /tmp/workspace/out and create a proper bugfix/<branchname> and pull request." > /etc/motd

WORKDIR /root