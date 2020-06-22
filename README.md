# plancosts

Get costs for a Terraform project.

Currently this supports the following On-Demand pricing for the following AWS resources:
 * `aws_instance`
 * `aws_ebs_volume`
 * `aws_ebs_snapshot`
 * `aws_ebs_snapshot_copy`
 * `aws_autoscaling_group`
 * `aws_elb`
 * `aws_lb`
 * `aws_nat_gateway`

This does not supports estimates for any costs that are not specified in the Terraform configuration, e.g. S3 storage costs, data out costs, etc.

This is an early stage project, pull requests to add resources/fix bugs are welcome.

## Table of Contents

* [Installation](#installation)
* [Usage](#usage)
* [Development](#development)
* [Contributing](#contributing)
* [License](#license)

## Installation

To download the latest release:

TODO: add release
```
curl --silent --location "https://github.com/aliscott/eksctl/plancosts/latest/download/plancosts_$(uname -s)_amd64.tar.gz" | tar xz -C /tmp
sudo mv /tmp/plancosts /usr/local/bin
```

## Usage

By default prices are retrieved from a [<TODO link to price list API repo>] deployed at <TODO once deployed>. This is running on minimal infrastructure so is not guaranteed to always be available.

You can also deploy the price list API yourself and specify it by setting the `PLANCOSTS_API_URL` env variable or passing the `--api-url` option.

Generate a cost breakdown from a Terraform directory:
```sh
plancosts --tfdir examples/terraform
```

Output the cost breakdown in JSON format:
```sh
plancosts --tfdir examples/terraform --output json
```

Generate a cost breakdown from a Terraform plan JSON file:
```sh
terraform plan -out plan.save examples/terraform
terraform show -json plan.save > plan.json

plancosts --tfplan-json plan.json
```

Generate a cost breakdown from a Terraform plan file:
```sh
terraform plan -out plan.save examples/terraform

plancosts --tfplan plan.save --tfdir examples/terraform
```

## Development

Install dependencies
```sh
make deps
```

Run the code
```sh
make run ARGS="--tfdir <Terraform Dir>"
```

Run tests:
```sh
make test
```

## Contributing

Pull requests are welcome. For major changes, please open an issue first to discuss what you would like to change.

## License

[ISC](https://choosealicense.com/licenses/isc/)
