<a href="https://www.infracost.io"><img src="https://raw.githubusercontent.com/infracost/infracost/master/assets/logo.svg" width=320 alt="Infracost logo" /></a>

<a href="https://www.infracost.io/community-chat"><img alt="Community Slack channel" src="https://img.shields.io/badge/chat-Slack-%234a154b"/></a>
<a href="https://www.infracost.io/docs/"><img alt="Docs" src="https://img.shields.io/badge/docs-blue"/></a>
<a href="https://github.com/infracost/infracost/actions?query=workflow%3AGo+branch%3Amaster"><img alt="Build Status" src="https://img.shields.io/github/workflow/status/infracost/infracost/Go/master"/></a>
<a href="https://hub.docker.com/r/infracost/infracost/tags"><img alt="Docker Image" src="https://img.shields.io/docker/cloud/build/infracost/infracost"/></a>
<a href="https://twitter.com/intent/tweet?text=Get%20cost%20estimates%20for%20cloud%20infrastructure%20in%20pull%20requests!&url=https://www.infracost.io&hashtags=cloud,cost,aws,IaC,terraform"><img alt="Tweet" src="https://img.shields.io/twitter/url/http/shields.io.svg?style=social"/></a>

Infracost shows cloud cost estimates for a Terraform project. It helps developers, devops and others to quickly see the cost breakdown and compare different options upfront.

<img src="https://raw.githubusercontent.com/infracost/infracost/master/assets/screenshot.png" width=600 alt="Example Infracost output" />

## Installation

1. Download and install the latest Infracost release

    macOS Homebrew:
    ```sh
    brew install infracost
    ```

    Linux/macOS manual:
    ```sh
    os=$(uname | tr [:upper:] [:lower:])
    curl -s -L https://github.com/infracost/infracost/releases/latest/download/infracost-$os-amd64.tar.gz | tar xz -C /tmp && \
    sudo mv /tmp/infracost-$os-amd64 /usr/local/bin/infracost
    ```

    Docker and Windows users see [here](https://www.infracost.io/docs/#installation).

2.	Use our free Cloud Pricing API by registering for an API key:
    ```sh
    infracost register
    ```

    If you prefer, you can run your own [Cloud Pricing API](https://www.infracost.io/docs/faq/#can-i-run-my-own-cloud-pricing-api).

3.  Run `infracost` using our example Terraform project to see how it works. You can also play with the `main.tf` file in the example:

    ```sh
    git clone https://github.com/infracost/example-terraform.git
    infracost --tfdir example-terraform/aws
    ```

Please **watch** this repo for new releases as we add new cloud resources every week or so.

## Basic usage

There are [4 usage methods](https://www.infracost.io/docs/#usage-methods) for Infracost depending on your use-case. The following is the default method. Point to the Terraform directory using `--tfdir` and pass any required Terraform flags using `--tfflags`. Internally Infracost runs Terraform `init`, `plan` and `show`; `init` requires cloud credentials to be set, e.g. via the usual `AWS_ACCESS_KEY_ID` environment variables. This method works with remote state too.
  ```sh
  infracost --tfdir /path/to/code --tfflags "-var-file=myvars.tfvars"
  ```

Checkout the [docs site](https://www.infracost.io/docs/) for additional usage options, including notes for [Terragrunt](https://www.infracost.io/docs/#terragrunt-users) and [Terraform Cloud](https://www.infracost.io/docs/#terraform-cloud-users) users.

## CI/CD integrations

The [Infracost GitHub Action](https://www.infracost.io/docs/integrations#github-action), [GitLab CI template](https://www.infracost.io/docs/integrations#gitlab-ci) or [CircleCI Orb](https://www.infracost.io/docs/integrations#circleci) can be used to automatically add a comment showing the cost estimate `diff` between a pull/merge request and the master branch. If you run into any issues with CI/CD integrations, please join our [community Slack channel](https://www.infracost.io/community-chat), we'd be happy to guide you through it.

<img src="https://raw.githubusercontent.com/infracost/infracost-gh-action/master/screenshot.png" width=600 alt="Example infracost diff usage" />

## Supported clouds and resources

Infracost supports over [50 AWS and Google resources](https://www.infracost.io/docs/supported_resources/), Azure is [planned](https://github.com/infracost/infracost/issues/64). The quickest way to find out if your Terraform resources are supported is to run Infracost with the `--show-skipped` option. This shows the unsupported resources, some of which might be free. Please watch this repo for new releases as we add new cloud resources every week or so.

See [this page](https://www.infracost.io/docs/usage_based_resources) for details on cost estimation of usage-based resources.

## Contributing

Issues and pull requests are welcome! For development details, see the [contributing](CONTRIBUTING.md) file. For major changes, please open an issue first to discuss what you would like to change.

[Join our community Slack channel](https://www.infracost.io/community-chat), we are a friendly bunch and happy to help you get started :)

We're looking for people to help us add new AWS and Google resources and are willing to pay for it. Please direct-message Ali Khajeh-Hosseini on our community Slack channel to find out more.

## License

[Apache License 2.0](https://choosealicense.com/licenses/apache-2.0/)
