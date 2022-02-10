#!/usr/bin/env sh
# This script is used in the README and https://www.infracost.io/docs/#quick-start
set -e

# check_sha is separated into a defined function so that we can
# capture the exit code effectively with `set -e` enabled
check_sha() {
  (
    cd /tmp/
    shasum -sc "$1"
  )

  return $?
}

os=$(uname | tr '[:upper:]' '[:lower:]')
arch=$(uname -m | tr '[:upper:]' '[:lower:]' | sed -e s/x86_64/amd64/)
if [ "$arch" = "aarch64" ]; then
  arch="arm64"
fi

url="https://infracost.io/downloads/latest"
tar="infracost-$os-$arch.tar.gz"
echo "Downloading latest release of infracost-$os-$arch..."
curl -sL "$url/$tar" -o "/tmp/$tar"
echo

code=$(curl -s -L -o /dev/null -w "%{http_code}" "$url/$tar.sha256")
if [ "$code" != "404" ]; then
  if [ -x "$(command -v shasum)" ]; then
    echo "Validating checksum for infracost-$os-$arch..."
    curl -sL "$url/$tar.sha256" -o "/tmp/$tar.sha256"

    if ! check_sha "$tar.sha256"; then
      echo
      echo "Installation checksum failed! This could be a security issue so please email hello@infracost.io for help."
      exit 1
    fi

    rm "/tmp/$tar.sha256"
  else
    echo "Skipping checksum validation as the shasum command could not be found, no action needed."
  fi
  echo
fi

tar xzf "/tmp/$tar" -C /tmp
rm "/tmp/$tar"

echo "Moving /tmp/infracost-$os-$arch to /usr/local/bin/infracost (you might be asked for your password due to sudo)"
if [ -x "$(command -v sudo)" ]; then
  sudo mv "/tmp/infracost-$os-$arch" "/usr/local/bin/infracost"
else
  mv "/tmp/infracost-$os-$arch" "/usr/local/bin/infracost"
fi
echo
echo "Completed installing $(infracost --version)"
