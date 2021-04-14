#!/usr/bin/env sh
# This script is used in the README and https://www.infracost.io/docs/#installation
set -e

os=$(uname | tr '[:upper:]' '[:lower:]')
arch=$(uname -m | tr '[:upper:]' '[:lower:]' | sed -e s/x86_64/amd64/)
echo "Downloading latest release of infracost-$os-$arch..."
curl -sL https://github.com/infracost/infracost/releases/latest/download/infracost-$os-$arch.tar.gz | tar xz -C /tmp
echo
echo "Moving /tmp/infracost-$os-$arch to /usr/local/bin/infracost (you might be asked for your password due to sudo)"
sudo mv /tmp/infracost-$os-$arch /usr/local/bin/infracost
echo
echo "Completed installing $(infracost --version)"
