FROM golang:1.25.4 AS builder

ARG ARCH=linux
ARG DEFAULT_TERRAFORM_VERSION=0.15.5
ARG TERRAGRUNT_VERSION=0.31.8

# Set Environment Variables
SHELL ["/bin/bash", "-c"]
ENV HOME=/app
ENV CGO_ENABLED=0

# Install Packages
RUN apt-get update -q && apt-get -y install unzip && rm -rf /var/lib/apt/lists/*

# Install latest of each Terraform version after 0.12 as we don't support versions before that
RUN AVAILABLE_TERRAFORM_VERSIONS="0.12.31 0.13.7 0.14.11 ${DEFAULT_TERRAFORM_VERSION} 1.0.2 1.0.10" && \
    for VERSION in ${AVAILABLE_TERRAFORM_VERSIONS}; do \
    wget -q https://releases.hashicorp.com/terraform/${VERSION}/terraform_${VERSION}_linux_amd64.zip && \
    wget -q https://releases.hashicorp.com/terraform/${VERSION}/terraform_${VERSION}_SHA256SUMS && \
    sed -n "/terraform_${VERSION}_linux_amd64.zip/p" terraform_${VERSION}_SHA256SUMS | sha256sum -c && \
    unzip terraform_${VERSION}_linux_amd64.zip -d /tmp && \
    mv /tmp/terraform /usr/bin/terraform_${VERSION} && \
    chmod +x /usr/bin/terraform_${VERSION} && \
    rm terraform_${VERSION}_linux_amd64.zip && \
    rm terraform_${VERSION}_SHA256SUMS; \
    done && \
    ln -s /usr/bin/terraform_0.12.31 /usr/bin/terraform_0.12 && \
    ln -s /usr/bin/terraform_0.13.7 /usr/bin/terraform_0.13 && \
    ln -s /usr/bin/terraform_0.14.11 /usr/bin/terraform_0.14 && \
    ln -s /usr/bin/terraform_1.0.10 /usr/bin/terraform_1.0 && \
    ln -s /usr/bin/terraform_${DEFAULT_TERRAFORM_VERSION} /usr/bin/terraform_0.15 && \
    ln -s /usr/bin/terraform_${DEFAULT_TERRAFORM_VERSION} /usr/bin/terraform

# Install Terragrunt
RUN wget -q https://github.com/gruntwork-io/terragrunt/releases/download/v$TERRAGRUNT_VERSION/terragrunt_linux_amd64
RUN mv terragrunt_linux_amd64 /usr/bin/terragrunt && \
    chmod +x /usr/bin/terragrunt

WORKDIR /app

# Build Application
COPY . .
RUN NO_DIRTY=true make build
RUN chmod +x /app/build/infracost

# Application
FROM alpine:3.16 AS app
# Tools needed for running diffs in CI integrations
RUN apk --no-cache add ca-certificates openssl openssh-client curl git bash

# The jq package provided by alpine:3.15 (jq 1.6-rc1) is flagged as a
# high severity vulnerability, so we install the latest release ourselves
# Reference: https://nvd.nist.gov/vuln/detail/CVE-2016-4074 (this is present on jq-1.6-rc1 as well)
RUN \
    # Install jq-1.6 (final release)
    curl -s -L -o /tmp/jq https://github.com/stedolan/jq/releases/download/jq-1.6/jq-linux64 && \
    mv /tmp/jq /usr/local/bin/jq && \
    chmod +x /usr/local/bin/jq

WORKDIR /root/
# Scripts are used by CI integrations and other use-cases
COPY scripts /scripts
COPY --from=builder /usr/bin/terraform* /usr/bin/
COPY --from=builder /usr/bin/terragrunt /usr/bin/
COPY --from=builder /app/build/infracost /usr/bin/

ENTRYPOINT [ "infracost" ]
