<img src="https://raw.githubusercontent.com/infracost/infracost/master/assets/logo.svg" width=320 alt="Infracost logo" />

<a href="https://docs.infracost.io"><img alt="Docs" src="https://img.shields.io/badge/docs-blue"/></a>
<a href="https://discord.gg/Cu9ftEg"><img alt="Discord Chat" src="https://img.shields.io/discord/746703155953270794.svg"/></a>
<a href="https://github.com/infracost/infracost/actions?query=workflow%3AGo+branch%3Amaster"><img alt="Build Status" src="https://img.shields.io/github/workflow/status/infracost/infracost/Go/master"/></a>
<a href="https://hub.docker.com/r/infracost/infracost/tags"><img alt="Docker Image" src="https://img.shields.io/docker/cloud/build/infracost/infracost"/></a>

Infracost shows hourly and monthly cost estimates for a Terraform project. This helps developers, DevOps et al. quickly see the cost breakdown and compare different deployment options upfront.

<img src="https://raw.githubusercontent.com/infracost/infracost/master/assets/screenshot.png" width=600 alt="Example Infracost output" />

## Table of Contents

**Checkout the [docs site](https://docs.infracost.io) for detailed usage options, supported resources and more information.**

* [Installation](#installation)
* [Usage](#basic-usage)
* [Development](#development)
* [Contributing](#contributing)

## Installation

To download and install the latest release:

```sh
curl --silent --location "https://github.com/infracost/infracost/releases/latest/download/infracost-$(uname -s)-amd64.tar.gz" | tar xz -C /tmp
sudo mv /tmp/infracost-$(uname -s | tr '[:upper:]' '[:lower:]')-amd64 /usr/local/bin/infracost
```

## Basic usage

Generate a cost breakdown from a Terraform directory:
```sh
infracost --tfdir examples/small_terraform
```

Check the [docs site](https://docs.infracost.io) for more details.

The [Infracost GitHub action](https://github.com/marketplace/actions/run-infracost) can be used to automatically add a PR comment showing the cost estimate `diff` between a pull request and the master branch whenever Terraform files change.

<img src="https://raw.githubusercontent.com/infracost/infracost-gh-action/master/screenshot.png" width=600 alt="Example infracost diff usage" />

## Development

Install dependencies:
```sh
make deps
```

Run the code:
```sh
make run ARGS="--tfdir <Terraform Dir>"
```

Run all tests:
```sh
make test
```

Exclude integration tests:
```sh
make test ARGS="-v -short"
```

Build:
```sh
make build
```

## Contributing

Pull requests are welcome. For major changes, please open an issue first to discuss what you would like to change.

## License

[Apache License 2.0](https://choosealicense.com/licenses/apache-2.0/)
