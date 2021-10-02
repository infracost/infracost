$outfile="infracost.zip"
$version=(Invoke-Webrequest https://api.github.com/repos/infracost/infracost/releases/latest|convertfrom-json).name

Write-Host "$(get-date) - downloading release $version"
Invoke-WebRequest -uri "https://github.com/infracost/infracost/releases/download/$($version)/infracost-windows-amd64.tar.gz" -OutFile $outfile
tar -xvf $outfile -C .\tools\

# removing the first v as chocolately doesnt like this version
$version=${version:1}
Write-Host "$(get-date) - packing"
choco pack --version $version

Get-ChildItem *.nupkg
Write-Host "$(get-date) - Pushing to Chocolatey Feed"
choco push $package.Name -s https://push.chocolatey.org/ --api-key=$env:CHOCOPUSHKEY