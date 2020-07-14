FROM alpine:3.12

# GitHub actions don't support build-args (https://github.community/t/feature-request-build-args-support-in-docker-container-actions/16846/4)
# so using ENV might help people who need to fork/change it
ENV TERRAFORM_VERSION=0.12.28 \
  TERRAFORM_BINARY=terraform \
  INFRACOST_VERSION=latest

RUN apk --update --no-cache add ca-certificates openssl sudo curl && \
  wget -O terraform.zip "https://releases.hashicorp.com/terraform/${TERRAFORM_VERSION}/terraform_${TERRAFORM_VERSION}_linux_amd64.zip" && \
  unzip terraform.zip -d /bin && \
  rm -rf terraform.zip /var/cache/apk/*

RUN curl --silent --location https://github.com/aliscott/infracost/releases/${INFRACOST_VERSION}/download/infracost-linux-amd64.tar.gz | tar xz -C /tmp
RUN mv /tmp/infracost-linux-amd64 /usr/local/bin/infracost

COPY entrypoint.sh /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
