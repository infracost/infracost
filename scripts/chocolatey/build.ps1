$outfile="infracost.tar.gz"
$version=(Invoke-Webrequest -UseBasicParsing https://api.github.com/repos/infracost/infracost/releases/latest|convertfrom-json).name

$uri="https://github.com/infracost/infracost/releases/download/$($version)/infracost-windows-amd64.tar.gz"
Write-Host "$(get-date) - downloading release $version from $uri"
Invoke-WebRequest -OutFile $outfile $uri
mkdir -Force tools
tar -xvf $outfile -C tools

# removing the leading 'v' to get a valid nupkg version number
$version=$version.Substring(1)
Write-Host "$(get-date) - packing"
choco pack --version $version

Get-ChildItem *.nupkg
Write-Host "$(get-date) - Pushing to Chocolatey Feed"
choco push $package.Name -s https://push.chocolatey.org/ --api-key=$env:CHOCOPUSHKEY