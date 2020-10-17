<a href="https://www.infracost.io"><img src="https://raw.githubusercontent.com/infracost/infracost/master/assets/logo.svg" width=320 alt="Infracost logo" /></a>

<a href="https://www.infracost.io/docs/"><img alt="Docs" src="https://img.shields.io/badge/docs-blue"/></a>
<a href="https://discord.gg/rXCTaH3"><img alt="Discord Chat" src="https://img.shields.io/discord/746703155953270794.svg"/></a>
<a href="https://github.com/infracost/infracost/actions?query=workflow%3AGo+branch%3Amaster"><img alt="Build Status" src="https://img.shields.io/github/workflow/status/infracost/infracost/Go/master"/></a>
<a href="https://hub.docker.com/r/infracost/infracost/tags"><img alt="Docker Image" src="https://img.shields.io/docker/cloud/build/infracost/infracost"/></a>
<a href="https://twitter.com/intent/tweet?text=Get%20cost%20estimates%20for%20cloud%20infrastructure%20in%20pull%20requests!&url=https://www.infracost.io&hashtags=cloud,cost,aws,IaC,terraform"><img alt="Tweet" src="https://img.shields.io/twitter/url/http/shields.io.svg?style=social"/></a>

Infracost shows hourly and monthly cost estimates for a Terraform project. This helps developers, DevOps et al. quickly see the cost breakdown and compare different deployment options upfront.

<img src="https://raw.githubusercontent.com/infracost/infracost/master/assets/screenshot.png" width=600 alt="Example Infracost output" />

## Table of Contents

**Checkout the [docs site](https://www.infracost.io/docs/) for detailed usage options, supported resources and more information.**

* [Installation](#installation)
* [Usage](#basic-usage)
* [Development](#development)
* [Contributing](#contributing)

<!-- NOTE: When updated also update https://github.com/infracost/docs/blob/master/docs/getting_started.md#installation with the same content -->
## Installation

1. Download and install the latest Infracost release

    Linux:
    ```sh
    curl -s -L https://github.com/infracost/infracost/releases/latest/download/infracost-linux-amd64.tar.gz | tar xz -C /tmp && \
    sudo mv /tmp/infracost-linux-amd64 /usr/local/bin/infracost
    ```

    macOS (Homebrew):
    ```sh
    brew install infracost
    ```

    macOS (manual):
    ```sh
    curl -s -L https://github.com/infracost/infracost/releases/latest/download/infracost-darwin-amd64.tar.gz | tar xz -C /tmp && \
    sudo mv /tmp/infracost-darwin-amd64 /usr/local/bin/infracost
    ```

    Docker:
    ```sh
    docker pull infracost/infracost
    docker run --rm \
      -e INFRACOST_API_KEY=see_following_step_on_how_to_get_this \
      -e AWS_ACCESS_KEY_ID=$AWS_ACCESS_KEY_ID \
      -e AWS_SECRET_ACCESS_KEY=$AWS_SECRET_ACCESS_KEY \
      -v $PWD/:/code/ infracost/infracost --tfdir /code/
      # add other required flags for infracost or envs for Terraform
    ```

2.	Use our free hosted API for cloud prices by registering for an API key:
    ```sh
    infracost register
    ```

    Alternatively you can run your [own pricing API](https://github.com/infracost/cloud-pricing-api) and set the `INFRACOST_PRICING_API_ENDPOINT` environment variable to point to it.

    The `INFRACOST_API_KEY` environment variable can be used to set to the API key in CI systems.

## Basic usage

Generate a cost breakdown from a Terraform directory:
```sh
infracost --tfdir /path/to/code --tfflags "-var-file=myvars.tfvars"
```

Check the [docs site](https://www.infracost.io/docs/) for more details.

The [Infracost GitHub action](https://github.com/marketplace/actions/run-infracost) can be used to automatically add a PR comment showing the cost estimate `diff` between a pull request and the master branch whenever Terraform files change.

<img src="https://raw.githubusercontent.com/infracost/infracost-gh-action/master/screenshot.png" width=600 alt="Example infracost diff usage" />

## Development

Install Go dependencies:
```sh
make deps
```

Install latest version of terraform-provider-infracost. If you want to use a local development version see [#using-a-local-version-of-terraform-provider-infracost](#using-a-local-version-of-terraform-provider-infracost)
```sh
make install_provider
```

Get an API key.
```sh
make run ARGS="register"
```
Alternatively checkout and run the [cloud-pricing-api](https://github.com/infracost/cloud-pricing-api) and set the `INFRACOST_PRICING_API_ENDPOINT` environment variable to point to it.

Add the API key to your `.env.local` file:
```
cat <<EOF >> .env.local
INFRACOST_API_KEY=XXX
EOF
```

Run the code:
```sh
make run ARGS="--tfdir examples/terraform"
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

### Using a local version of terraform-provider-infracost

To use a local development version of terraform-provider-infracost

1. Fork/clone the [terraform-provider-infracost repository](https://github.com/infracost/terraform-provider-infracost)

2. Inside the directory that you cloned the repository run the following to install the local version in your `~/.terraform.d/plugins` directory:
  ```sh
  make install
  ```

## Contributing

Pull requests are welcome! For more info, see the [CONTRIBUTING](CONTRIBUTING.md) file. For major changes, please open an issue first to discuss what you would like to change.

Join our chat, we are a friendly bunch and happy to help you get started :) https://discord.gg/rXCTaH3

## License

[Apache License 2.0](https://choosealicense.com/licenses/apache-2.0/)
