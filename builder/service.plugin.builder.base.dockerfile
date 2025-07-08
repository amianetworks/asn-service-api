# Copyright 2025 Amiasys Corporation and/or its affiliates. All rights reserved.

# Build base build image for ASN Service Plugins

FROM ubuntu:24.04

WORKDIR /asn-service

## Install critical dependencies in one layer with noninteractive mode
RUN apt update && \
    apt install -y build-essential wget git ca-certificates gnupg2 protobuf-compiler && \
    apt clean && \
    rm -rf /var/lib/apt/lists/*

## Install dpkg-dev
RUN apt install -y dpkg-dev && \
    apt clean && \
    rm -rf /var/lib/apt/lists/*

# Install Go
RUN wget -q https://go.dev/dl/go1.24.0.linux-amd64.tar.gz && \
    tar -C /etc -xzf go1.24.0.linux-amd64.tar.gz && \
    rm -f go1.24.0.linux-amd64.tar.gz
ENV PATH="${PATH}:/etc/go/bin"
ENV GOPROXY="https://goproxy.io,direct"
#ENV GOPATH=/go
#ENV GOCACHE=${GOPATH}/.cache
#ENV GOMODCACHE=${GOPATH}/pkg/mod

# Configure SSH for private GitHub repositories
#RUN git config --global --add url."git@github.com:".insteadOf "https://github.com/" && \
#    mkdir -p /root/.ssh && \
#    chmod 700 /root/.ssh
#COPY ./ssh /root/.ssh
#RUN chmod 400 /root/.ssh/id_rsa
#ENV GOPRIVATE="github.com/amianetworks/*"

# Copy project files
COPY . .

# Run build.so once to get all Go packages downloaded.
RUN make -f make/internal.mk build.so

# Clean up the workdir for later builds.
WORKDIR /
RUN rm -rf /asn-service

# Default 
CMD ["bash"]
