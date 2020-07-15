# infracost

Get cost hourly and monthly estimates for a Terraform project. Helps you quickly see the cost breakdown and compare different deployment options upfront.

<img src="examples/screenshot.png" width=557 alt="Example infracost output" />

The [infracost GitHub action](https://github.com/aliscott/infracost-gh-action) can be used to automatically add a PR comment showing the cost estimate `diff` between a pull request and the master branch whenever a `.tf` file changes.

<img src="https://raw.githubusercontent.com/aliscott/infracost-gh-action/master/diff_screenshot.png" width=557 alt="Example infracost diff usage" />

Currently this supports the following On-Demand pricing for the following AWS resources:
 * `aws_instance`
 * `aws_ebs_volume`
 * `aws_ebs_snapshot`
 * `aws_ebs_snapshot_copy`
 * `aws_autoscaling_group`
 * `aws_db_instance`
 * `aws_elb`
 * `aws_lb`
 * `aws_nat_gateway`

This does not yet support estimates for:
  * any costs that are not specified in the Terraform configuration, e.g. S3 storage costs, data out costs.
  * Non-Linux EC2 instances such as Windows and RHEL, a lookup is needed to find the OS of AMIs.
  * Any non On-Demand pricing, such as Reserved Instances, Spot Instances.

This is an early stage project, pull requests to add resources/fix bugs are welcome.

## Table of Contents

* [Installation](#installation)
* [Usage](#usage)
* [Development](#development)
* [Contributing](#contributing)
* [License](#license)

## Installation

To download and install the latest release:

```sh
curl --silent --location "https://github.com/aliscott/infracost/releases/latest/download/infracost-$(uname -s)-amd64.tar.gz" | tar xz -C /tmp
sudo mv /tmp/infracost-$(uname -s)-amd64 /usr/local/bin/infracost
```

## Usage

Prices are retrieved using [https://github.com/aliscott/aws-prices-graphql](https://github.com/aliscott/aws-prices-graphql). There is a demo version of that service deployed at [https://aws-prices-graphql.alistair.scot/graphql](https://aws-prices-graphql.alistair.scot/graphql), which `infracost` uses by default. This is running on minimal infrastructure so is not guaranteed to always be available.

You can run `infracost` in your terraform directories without worrying about security or privacy issues as no terraform secrets/tags/IDs etc are sent to the pricing service (only generic price-related attributes are used). Also, do not be alarmed by seeing the `terraform init` in output, no changes are made to your terraform or cloud resources. As a security precaution, read-only AWS IAM creds can be used.

You can also deploy the price list API yourself and specify it by setting the `infracost_API_URL` env variable or passing the `--api-url` option.

Generate a cost breakdown from a Terraform plan JSON file:
```sh
infracost --tfjson examples/terraform/plan.json
```

To generate the plan JSON file you can run:
```
terraform plan -out plan.save examples/terraform
terraform show -json plan.save > examples/terraform/plan.json
```

Generate a cost breakdown from a Terraform directory:
```sh
infracost --tfdir examples/terraform
```

Generate a cost breakdown from a Terraform plan file:
```sh
infracost --tfplan plan.save --tfdir examples/terraform
```

To change the path to your `terraform` binary you can set the `TERRAFORM_BINARY` env variable:
```sh
TERRAFORM_BINARY=~/bin/terraform_0.13 infracost --tfdir examples/terraform_0.13
```

## Development

Install dependencies:
```sh
make deps
```

Run the code:
```sh
make run ARGS="--tfdir <Terraform Dir>"
```

Run tests:
```sh
make test
```

Build:
```sh
make build
```

## Contributing

Pull requests are welcome. For major changes, please open an issue first to discuss what you would like to change.

## License

[MIT](https://choosealicense.com/licenses/mit/)
