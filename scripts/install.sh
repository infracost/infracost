#!/usr/bin/env sh
# This script is used in the README and https://www.infracost.io/docs/#quick-start
set -e

os=$(uname | tr '[:upper:]' '[:lower:]')
arch=$(uname -m | tr '[:upper:]' '[:lower:]' | sed -e s/x86_64/amd64/)
if [ "$arch" = "aarch64" ]; then
  arch="arm64"
fi

url="https://infracost.io/downloads/latest"
tar="infracost-$os-$arch.tar.gz"
echo "Downloading latest release of infracost-$os-$arch..."
curl -sL $url/$tar -o /tmp/$tar
echo

code=$(curl -s -L -o /dev/null -w "%{http_code}" $url/$tar.sha256)
if [ "$code" = "404" ]; then
  echo "Checksum for infracost-$os-$arch release not found, skipping checksum validation"
else
  if [ -x "$(command -v shasum)" ]; then
    echo "Validating checksum for infracost-$os-$arch..."
    curl -sL $url/$tar.sha256 -o /tmp/$tar.sha256
    (cd /tmp/;shasum -sqc $tar.sha256)
    rm /tmp/$tar.sha256
  else
    echo "The shasum command could not be found, skipping checksum validation"
  fi
fi

tar xzf /tmp/$tar -C /tmp
rm /tmp/$tar

echo
echo "Moving /tmp/infracost-$os-$arch to /usr/local/bin/infracost (you might be asked for your password due to sudo)"
if [ -x "$(command -v sudo)" ]; then
  sudo mv /tmp/infracost-$os-$arch /usr/local/bin/infracost
else
  mv /tmp/infracost-$os-$arch /usr/local/bin/infracost
fi
echo
echo "Completed installing $(infracost --version)"
