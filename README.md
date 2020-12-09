<a href="https://www.infracost.io"><img src="https://raw.githubusercontent.com/infracost/infracost/master/assets/logo.svg" width=320 alt="Infracost logo" /></a>

<a href="https://www.infracost.io/docs/"><img alt="Docs" src="https://img.shields.io/badge/docs-blue"/></a>
<a href="https://www.infracost.io/community-chat"><img alt="Community Slack channel" src="https://img.shields.io/badge/chat-Slack-%234a154b"/></a>
<a href="https://github.com/infracost/infracost/actions?query=workflow%3AGo+branch%3Amaster"><img alt="Build Status" src="https://img.shields.io/github/workflow/status/infracost/infracost/Go/master"/></a>
<a href="https://hub.docker.com/r/infracost/infracost/tags"><img alt="Docker Image" src="https://img.shields.io/docker/cloud/build/infracost/infracost"/></a>
<a href="https://twitter.com/intent/tweet?text=Get%20cost%20estimates%20for%20cloud%20infrastructure%20in%20pull%20requests!&url=https://www.infracost.io&hashtags=cloud,cost,aws,IaC,terraform"><img alt="Tweet" src="https://img.shields.io/twitter/url/http/shields.io.svg?style=social"/></a>

Infracost shows hourly and monthly cost estimates for a Terraform project. This helps developers, DevOps et al. quickly see the cost breakdown and compare different deployment options upfront.

<img src="https://raw.githubusercontent.com/infracost/infracost/master/assets/screenshot.png" width=600 alt="Example Infracost output" />

## Table of Contents

**Checkout the [docs site](https://www.infracost.io/docs/) for additional usage options.**

* [Installation](#installation)
* [Usage](#usage-methods)
* [Supported clouds and resources](https://www.infracost.io/docs/supported_resources/)
* [Development](#development)
* [Contributing (we'll pay you!)](#contributing)

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

    The `INFRACOST_API_KEY` environment variable can be used to set the API key in CI systems.
    If you prefer, you can run your own [pricing API](https://www.infracost.io/docs/faq/#can-i-run-my-own-pricing-api).

3.  Run `infracost` using our example Terraform project to see how it works. You can also play with the `main.tf` file in the example:

    ```sh
    git clone https://github.com/infracost/example-terraform.git
    infracost --tfdir example-terraform/aws
    ```

## Basic usage

There are [4 usage methods](https://www.infracost.io/docs/#usage-methods) for Infracost depending on your use-case. The following is the default method. Point to the Terraform directory using `--tfdir` and pass any required Terraform flags using `--tfflags`. Internally Infracost runs Terraform `init`, `plan` and `show`; `init` requires cloud credentials to be set, e.g. via the usual `AWS_ACCESS_KEY_ID` environment variables. This method works with remote state too.
  ```sh
  infracost --tfdir /path/to/code --tfflags "-var-file=myvars.tfvars"
  ```

The [Infracost GitHub Action](https://www.infracost.io/docs/integrations#github-action) or [GitLab CI template](https://www.infracost.io/docs/integrations#gitlab-ci) can be used to automatically add a PR comment showing the cost estimate `diff` between a pull/merge request and the master branch.

<img src="https://raw.githubusercontent.com/infracost/infracost-gh-action/master/screenshot.png" width=600 alt="Example infracost diff usage" />

## Development

Install Go dependencies:
```sh
make deps
```

Install latest version of terraform-provider-infracost. If you want to use a local development version, see [this](#using-a-local-version-of-terraform-provider-infracost)
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

[Join our community Slack channel](https://www.infracost.io/community-chat), we are a friendly bunch and happy to help you get started :)

We're looking for people to help us add new AWS and Google resources and are willing to pay for that. Please direct-message Ali Khajeh-Hosseini on our community Slack channel to find out more!

## License

[Apache License 2.0](https://choosealicense.com/licenses/apache-2.0/)
