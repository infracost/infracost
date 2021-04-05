#!/usr/bin/env sh

set -e

os=$(uname | tr '[:upper:]' '[:lower:]')
arch=$(uname -m | tr '[:upper:]' '[:lower:]' | sed -e s/x86_64/amd64/)
curl -s -L https://github.com/infracost/infracost/releases/latest/download/infracost-$os-$arch.tar.gz | tar xz -C /tmp
sudo mv /tmp/infracost-$os-$arch /usr/local/bin/infracost
