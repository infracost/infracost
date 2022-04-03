<p align="center">
<a href="https://www.infracost.io"><img src=".github/assets/logo.svg" alt="Infracost breakdown command" width="300" /></a>

<p align="center">Infracost shows cloud cost estimates for Terraform. It lets DevOps, SRE and engineers see a cost breakdown and understand costs <b>before making changes</b>, either in the terminal or pull requests.</p>
</p>
<p align="center">
<a href="https://www.infracost.io/docs/"><img alt="Docs" src="https://img.shields.io/badge/docs-get%20started-brightgreen"/></a>
<img alt="Docker pulls" src="https://img.shields.io/docker/pulls/infracost/infracost"/>
<a href="https://www.infracost.io/community-chat"><img alt="Community Slack channel" src="https://img.shields.io/badge/chat-slack-%234a154b"/></a>
<a href="https://twitter.com/intent/tweet?text=Get%20cost%20estimates%20for%20Terraform%20in%20pull%20requests!&url=https://www.infracost.io&hashtags=cloud,cost,terraform"><img alt="tweet" src="https://img.shields.io/twitter/url/http/shields.io.svg?style=social"/></a>
</p>

## Get started

Follow our [**quick start guide**](https://www.infracost.io/docs/#quick-start) to get started with the CLI 🚀

Infracost has many [CI/CD integrations](https://www.infracost.io/docs/integrations/cicd/) so you can easily post cost estimates in pull requests. This provides your team with a safety net as people can discuss costs as part of the workflow.

#### Post cost estimates in pull requests

<img src=".github/assets/github_actions_screenshot.png" alt="Infracost in GitHub Actions" width=700 />

#### Show `diff` of monthly costs between current and planned state in CLI

<img src=".github/assets/diff_screenshot.png" alt="Infracost diff command" width=600 />

#### Show `breakdown` of costs in CLI

<img src=".github/assets/breakdown_screenshot.png" alt="Infracost breakdown command" width=600 />

## Supported clouds and resources

Infracost supports over **230** Terraform resources across [AWS](https://www.infracost.io/docs/supported_resources/aws), [Azure](https://www.infracost.io/docs/supported_resources/azure) and [Google](https://www.infracost.io/docs/supported_resources/google). Other IaC tools, such as [Pulumi](https://github.com/infracost/infracost/issues/187), [AWS CloudFormation/CDK](https://github.com/infracost/infracost/issues/190) and [Azure ARM/Bicep](https://github.com/infracost/infracost/issues/812) are on our roadmap.

Infracost can also estimate [usage-based resources](https://www.infracost.io/docs/usage_based_resources) such as AWS S3 or Lambda!

## Community and contributing

Join our [community Slack channel](https://www.infracost.io/community-chat) to learn more about cost estimation, Infracost, and to connect with other users and contributors.

We ❤️ contributions big or small. For development details, see the [contributing](CONTRIBUTING.md) guide. For major changes, including CLI interface changes, please open an issue first to discuss what you would like to change.

Thanks to all the people who have contributed, including bug reports, code, feedback and suggestions!

<a href="https://github.com/infracost/infracost/graphs/contributors">
  <img src="https://contrib.rocks/image?repo=infracost/infracost" />
</a>

## License

[Apache License 2.0](https://choosealicense.com/licenses/apache-2.0/)
