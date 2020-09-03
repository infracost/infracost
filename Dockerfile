FROM golang:1.15 as builder

ARG ARCH=linux
ARG TERRAFORM_VERSION=0.12.25

# Set Environment Variables
SHELL ["/bin/bash", "-c"]
ENV HOME /app
ENV CGO_ENABLED 0
ENV GOOS linux

# Install Packages
RUN wget -q https://releases.hashicorp.com/terraform/${TERRAFORM_VERSION}/terraform_${TERRAFORM_VERSION}_${ARCH}_amd64.zip
RUN apt-get update && apt-get -y install zip -y
RUN unzip terraform*.zip && \
    mv terraform /usr/local/bin && \
    chmod +x /usr/local/bin/terraform

# Build Application
WORKDIR /app
COPY . .
RUN make deps
RUN make build

# Application
FROM alpine:3.12 as app
ARG TERRAFORM_VERSION=0.12.25
RUN apk --no-cache add ca-certificates terraform=${TERRAFORM_VERSION}-r0
WORKDIR /root/
COPY --from=builder /app/build/infracost /usr/local/bin/infracost
RUN chmod +x /usr/local/bin/infracost

ENTRYPOINT [ "infracost" ]
CMD [ "--help" ]
