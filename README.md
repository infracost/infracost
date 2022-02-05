[![Infracost logo](.github/assets/logo.svg)](https://www.infracost.io)


<a href="https://www.infracost.io/docs/"><img alt="Docs" src="https://img.shields.io/badge/docs-Get%20started-brightgreen"/></a>
<a href="https://www.infracost.io/community-chat"><img alt="Community Slack channel" src="https://img.shields.io/badge/chat-Slack-%234a154b"/></a>
<a href="https://github.com/infracost/infracost/actions?query=workflow%3AGo+branch%3Amaster"><img alt="Build Status" src="https://img.shields.io/github/workflow/status/infracost/infracost/Go/master"/></a>
<a href="https://www.infracost.io/docs/integrations/cicd/#docker-images"><img alt="Docker Image" src="https://img.shields.io/docker/cloud/build/infracost/infracost"/></a>
<a href="https://twitter.com/intent/tweet?text=Get%20cost%20estimates%20for%20Terraform%20in%20pull%20requests!&url=https://www.infracost.io&hashtags=cloud,cost,terraform"><img alt="Tweet" src="https://img.shields.io/twitter/url/http/shields.io.svg?style=social"/></a>

Infracost shows cloud cost estimates for Terraform. It enables DevOps, SRE and engineers to see a cost breakdown and understand costs **before making changes**, either in the terminal or pull requests. This provides your team with a safety net as people can discuss costs as part of the workflow.

<img src="https://raw.githubusercontent.com/infracost/actions/master/.github/assets/screenshot.png" alt="Infracost in GitHub Actions" width=800 />

## Quick start

### 1. Install Infracost

Assuming [Terraform](https://www.terraform.io/downloads.html) is already installed, get the latest Infracost release:

macOS Homebrew:
```sh
brew install infracost
```

Linux/macOS manual download:
```sh
# Downloads the CLI based on your OS/arch and puts it in /usr/local/bin
curl -fsSL https://raw.githubusercontent.com/infracost/infracost/master/scripts/install.sh | sh
```

Docker and Windows users see [here](https://www.infracost.io/docs/#quick-start).

### 2. Get API key

Register for a free API key, which is used by the CLI to retrieve prices from our Cloud Pricing API, e.g. get prices for instance types. No cloud credentials or secrets are [sent](https://www.infracost.io/docs/faq/#what-data-is-sent-to-the-cloud-pricing-api) to the API and you can also [self-host](https://www.infracost.io/docs/cloud_pricing_api/self_hosted/) it.

```sh
infracost register
```

The key can be retrieved with `infracost configure get api_key`.

### 3. Run it

Infracost does not make any changes to your Terraform state or cloud resources. Run Infracost using our example Terraform project to see how it works. The [CLI commands](https://www.infracost.io/docs/features/cli_commands/) page describes the options for `--path`, which can point to a Terraform directory or plan JSON file.

```sh
git clone https://github.com/infracost/example-terraform.git
cd example-terraform/sample1

# Play with main.tf and re-run to compare costs
infracost breakdown --path .

# Show diff of monthly costs, edit the yml file and re-run to compare costs
infracost diff --path . --sync-usage-file --usage-file infracost-usage.yml
```

Screenshots of example outputs are [shown below](#cli-commands).

### 4. Add to CI/CD

Use our CI/CD integrations to add cost estimates to pull requests. This provides your team with a safety net as people can understand cloud costs upfront, and discuss them as part of your workflow.
- [GitHub Actions](https://www.infracost.io/docs/integrations/github_actions/)
- [GitLab CI](https://www.infracost.io/docs/integrations/gitlab_ci/)
- [Atlantis](https://www.infracost.io/docs/integrations/atlantis/)
- [Azure DevOps](https://www.infracost.io/docs/integrations/cicd/#azure-devops)
- [Terraform Cloud/Enterprise](https://www.infracost.io/docs/integrations/terraform_cloud_enterprise/)
- [Jenkins](https://www.infracost.io/docs/integrations/cicd/#jenkins)
- [Bitbucket Pipelines](https://www.infracost.io/docs/integrations/cicd/#bitbucket-pipelines)
- [CircleCI](https://www.infracost.io/docs/integrations/cicd/#circleci)

Other CI/CD systems can be supported using [our Docker images](https://www.infracost.io/docs/integrations/cicd/#docker-images). You can also setup [cost policies](https://www.infracost.io/docs/features/cost_policies/).

If you run into any issues, please join our [community Slack channel](https://www.infracost.io/community-chat), we'll help you very quickly 😄

## CLI commands

The `infracost` CLI has the following main commands, see [our docs](https://www.infracost.io/docs/features/cli_commands/) for the other commands:

#### Show full breakdown of costs

<img src=".github/assets/breakdown_screenshot.png" alt="Infracost breakdown command" width=600 />

#### Show diff of monthly costs between current and planned state

<img src=".github/assets/diff_screenshot.png" alt="Infracost diff command" width=600 />

## Supported clouds and resources

Infracost supports over **200** Terraform resources across [AWS](https://www.infracost.io/docs/supported_resources/aws), [Azure](https://www.infracost.io/docs/supported_resources/azure) and [Google](https://www.infracost.io/docs/supported_resources/google). Other IaC tools, such as [Pulumi](https://github.com/infracost/infracost/issues/187), [AWS CloudFormation/CDK](https://github.com/infracost/infracost/issues/190) and [Azure ARM/Bicep](https://github.com/infracost/infracost/issues/812) are on our roadmap.

See [this page](https://www.infracost.io/docs/usage_based_resources) for details on cost estimation of usage-based resources such as AWS Lambda or Google Cloud Storage.

## Contributing

Issues and pull requests are welcome! For development details, see the [contributing](CONTRIBUTING.md) guide. For major changes, including CLI interface changes, please open an issue first to discuss what you would like to change. [Join our community Slack channel](https://www.infracost.io/community-chat), we are a friendly bunch and happy to help you get started :)

Thanks to all the people who have contributed, including bug reports, code, feedback and suggestions!

<a href="https://github.com/infracost/infracost/graphs/contributors">
  <img src="https://contrib.rocks/image?repo=infracost/infracost" />
</a>

## License

[Apache License 2.0](https://choosealicense.com/licenses/apache-2.0/)
