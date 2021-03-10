#!/usr/bin/env bash

# This script runs infracost on all subfolders that have .tf files and outputs the combined
# diff using the `infracost output` command.
# You can customize it based on which folders it should exclude or how you run infracost.

# Find all subfolders that have .tf files, but exclude "modules" folders, can be customized
tfprojects=$(find . -type f -name '*.tf' | sed -E 's|/[^/]+$||' | grep -v modules | sort -u)

# Run infracost on the folders individually
while IFS= read -r tfproject; do
  echo "Running infracost diff for $tfproject"
  filename=$(echo $tfproject | sed 's:/:-:g' | cut -c3-)
  # TODO: customize to how you run infracost
  infracost breakdown --path $tfproject --format json > "$filename-infracost-out.json"
done <<< "$tfprojects"

# Run infracost output to merge the subfolder results
jsonfiles=($(find . -name "*-infracost-out.json" | tr '\n' ' '))
infracost output --format diff $(echo ${jsonfiles[@]/#/--path })

# Remove temp json files
rm $jsonfiles
