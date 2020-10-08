FROM golang:1.15 as builder

ARG ARCH=linux
ARG TERRAFORM_VERSION=0.13.4
ARG TERRAFORM_PROVIDER_INFRACOST_VERSION=latest

# Set Environment Variables
SHELL ["/bin/bash", "-c"]
ENV HOME /app
ENV CGO_ENABLED 0

# Install Packages
RUN wget -q https://releases.hashicorp.com/terraform/${TERRAFORM_VERSION}/terraform_${TERRAFORM_VERSION}_${ARCH}_amd64.zip
RUN apt-get update && apt-get -y install zip jq -y
RUN unzip terraform*.zip && \
    mv terraform /usr/local/bin && \
    chmod +x /usr/local/bin/terraform

WORKDIR /app
COPY scripts/install_provider.sh scripts/install_provider.sh
RUN scripts/install_provider.sh ${TERRAFORM_PROVIDER_INFRACOST_VERSION} /usr/local/bin/

# Build Application
COPY . .
RUN make deps
RUN make build

# Application
FROM alpine:3.12 as app
ARG TERRAFORM_VERSION=0.13.4
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /usr/local/bin/terraform* /usr/bin/
COPY --from=builder /usr/local/bin/terraform-provider-infracost* /usr/bin/
COPY --from=builder /app/build/infracost /usr/local/bin/infracost
RUN chmod +x /usr/local/bin/infracost
ENTRYPOINT [ "infracost" ]
CMD [ "--help" ]
