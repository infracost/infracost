#!/usr/bin/env bash

# This script runs infracost on all subfolders that have .tf files and outputs the combined
# results using the `infracost report` command. It also saves an infracost-report.html file.
# You can customize it based on which folders it should exclude or how you run infracost.

# Find all subfolders that have .tf files, but exclude "modules" folders, can be customized
tfprojects=$(find . -type f -name '*.tf' | sed -E 's|/[^/]+$||' | grep -v modules | sort -u)

# Run infracost on the folders individually
while IFS= read -r tfproject; do
  echo "Running infracost for $tfproject"
  cd $tfproject
  filename=$(echo $tfproject | sed 's:/:-:g' | cut -c3-)
  # TODO: customize to how you run infracost
  infracost --tfdir . --output json > "$filename-infracost-out.json"
  cd - > /dev/null
done <<< "$tfprojects"

# Run infracost report to merge the subfolder results
jsonfiles=$(find . -name "*-infracost-out.json")
infracost report --output html $(echo $jsonfiles | tr '\n' ' ') > report.html
infracost report --output table $(echo $jsonfiles | tr '\n' ' ')
echo "Also saved HTML report in infracost-report.html"

# Remove temp json files
rm $(echo $jsonfiles | tr '\n' ' ')
