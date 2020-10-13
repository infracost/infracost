#!/usr/bin/env bash

if ! command -v jq &> /dev/null
then
  echo "This script requires jq"
  exit
fi

version=${1:-latest}
if [ ! -z ${2+x} ]; then
  install_path=${2}
fi

# This isn't needed, but can be set if you hit rate limiting
if [ ! -z $GITHUB_ACCESS_TOKEN} ]; then
  github_headers="-H 'Authorization: token ${GITHUB_ACCESS_TOKEN}'"
fi

goos=$(go env GOOS)
goarch=$(go env GOARCH)

# Download the release assets from GitHub to a tmp directory
tmp_dir=$(mktemp)
if [ "$version" == "latest" ]; then
  resp=$(curl "${github_headers}" https://api.github.com/repos/infracost/terraform-provider-infracost/releases/latest)
  version=$(echo ${resp} | jq -r '.name')
  echo ${resp} | jq -r '.assets[] | select (.name | contains("'${goos}'_'${goarch}'")) | .browser_download_url' | wget -P ${tmp_dir} -i -
else
  curl "${github_headers}" https://api.github.com/repos/infracost/terraform-provider-infracost/releases \
      | jq -r '.[] | select (.name == "'${version}'") | .assets[] | select (.name | contains("${goos}_${goarch}")) | .browser_download_url' | wget -P ${tmp_dir} -i -
fi

# Unzip the release assets
zip=$(ls ${tmp_dir} | grep "terraform-provider-infracost.*\.zip")
unzip -od ${tmp_dir} ${tmp_dir}/${zip} && rm ${tmp_dir}/${zip}
binary=$(ls ${tmp_dir} | grep "terraform-provider-infracost")

# Strip any v prefix from the version
version=${version#"v"}

# If the install path isn't set then use the default Terraform plugin install paths
if [ -z ${install_path+x} ]; then
  plugin_root_path=${HOME}/.terraform.d/plugins
  install_path=${plugin_root_path}/registry.terraform.io/infracost/infracost/${version}/${goos}_${goarch}

  mkdir -p ${install_path}
  mv ${tmp_dir}/${binary} ${install_path}/${binary}

  # Add a link from the old-style plugin paths to maintain compatibility with older versions of Terraform
  mkdir -p ${plugin_root_path}/${goos}_${goarch}
  ln -sf ${install_path}/${binary} ${plugin_root_path}/${goos}_${goarch}/${binary}
# Otherwise use the user-specified install path
else
  mkdir -p ${install_path}
  mv ${tmp_dir}/${binary} ${install_path}/${binary}
fi
