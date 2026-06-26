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

# This script only installs versions <1.0.0; the CLI moved to https://github.com/infracost/cli.
# "latest" is resolved against this repo's releases rather than the infracost.io downloads endpoint,
# which now points to the new CLI.
version=${INFRACOST_VERSION:-latest}
if [ "$version" = "latest" ]; then
  echo "Fetching the latest release tag from infracost/infracost..."
  version=$(curl -sL https://api.github.com/repos/infracost/infracost/releases/latest | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
  if [ -z "$version" ]; then
    echo "Failed to fetch the latest release tag from GitHub. Please set INFRACOST_VERSION explicitly."
    exit 1
  fi
fi

major=$(echo "$version" | sed 's/^v//' | cut -d. -f1)
if echo "$major" | grep -qE '^[0-9]+$' && [ "$major" -ge 1 ]; then
  echo "Error: version $version is not supported by this script."
  echo "This script only installs versions <1.0.0 from the infracost/infracost repository."
  echo "For Infracost CLI >=2.0.0, see https://github.com/infracost/cli."
  exit 1
fi

url="https://infracost.io/downloads/${version}"
tar="infracost-$os-$arch.tar.gz"
echo "Downloading version ${version} of infracost-$os-$arch..."
curl -sL "$url/$tar" -o "/tmp/$tar"
echo

code=$(curl -s -L -o /dev/null -w "%{http_code}" "$url/$tar.sha256")
if [ "$code" = "404" ]; then
    echo "Skipping checksum validation as the sha for the release could not be found, no action needed."
else
  if [ -x "$(command -v shasum)" ]; then
    echo "Validating checksum for infracost-$os-$arch..."
    curl -sL "$url/$tar.sha256" -o "/tmp/$tar.sha256"

    if ! check_sha "$tar.sha256"; then
      echo
      read -r -p "Installation checksum failed. This could be a security issue. Would you like to continue? (y/n) " answer
      if [ "$answer" != "y" ]; then
        echo
        echo "Exiting, please email hello@infracost.io for help."
        exit 1
      fi
    fi

    rm "/tmp/$tar.sha256"
  else
    echo "Skipping checksum validation as the shasum command could not be found, no action needed."
  fi
fi
echo

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
