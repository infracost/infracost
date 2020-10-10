---
slug: supported_resources
title: Supported resources
---

Currently Infracost supports the following Terraform resources on AWS.

Support for the following is not currently included:
  * any costs that are not specified in the Terraform configuration, e.g. S3 storage costs, data out costs.
  * Any non On-Demand pricing, such as Reserved Instances.

| Terraform resource           | Notes |
| ---                          | ---   |
{{ range $key, $value := . }}{{ if not .NoPrice}}| `{{ $key }}` | {{ range $note := $value.Notes }} {{ $note }}<br /> {{ end }} |
{{ end }}{{ end }}

## The resource I want isn't supported

We're regularly adding support for new Terraform AWS resources - be sure to watch the repo for new [releases](https://github.com/infracost/infracost/releases)! We plan to add support for more cloud vendors ([GCP](https://github.com/infracost/infracost/issues/24), Azure) and other IaC tools (Pulumi) too.

You can help by:
1. [Creating an issue](https://github.com/infracost/infracost/issues/new) and mentioning the resource you need and a little about your use-case; we'll try to prioritize it depending on the community feedback.
2. [Contributing to Infracost](https://github.com/infracost/infracost#contributing), we're working on making it easier to add new resources. You can join our [Discord community](https://discord.gg/rXCTaH3) if you need help contributing.
