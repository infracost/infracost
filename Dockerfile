FROM golang:1.19 as builder

ARG ARCH=linux

# Set Environment Variables
SHELL ["/bin/bash", "-c"]
ENV HOME /app
ENV CGO_ENABLED 0

WORKDIR /app

# Build Application
COPY . .
RUN make deps
RUN NO_DIRTY=true make build
RUN chmod +x /app/build/infracost

# Application
FROM alpine:3.16 as app
# Tools needed for running diffs in CI integrations
RUN apk --no-cache add ca-certificates bash curl git openssl openssh-client

WORKDIR /root/

COPY --from=builder /app/build/infracost /usr/bin/

ENTRYPOINT [ "infracost" ]
