$outfile="infracost.zip"
$version=(Invoke-Webrequest https://api.github.com/repos/infracost/infracost/releases/latest|convertfrom-json).name

Write-Host "$(get-date) - downloading release $version"
Invoke-WebRequest -uri "https://github.com/bridgecrewio/yor/releases/download/$($version)/yor-$($version)-windows-amd64.zip" -OutFile $outfile
tar -xvf $outfile -C .\tools\

Write-Host "$(get-date) - packing"
choco pack --version $version

Get-ChildItem *.nupkg
Write-Host "$(get-date) - Pushing to Chocolatey Feed"
choco push $package.Name -s https://push.chocolatey.org/ --api-key=$env:CHOCOPUSHKEY