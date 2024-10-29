package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func getLambdaProvisionedConcurrencyConfigRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "aws_lambda_provisioned_concurrency_config",
		CoreRFunc: NewLambdaProvisionedConcurrencyConfig,
	}
}

func NewLambdaProvisionedConcurrencyConfig(d *schema.ResourceData) schema.CoreResource {
	region := d.Get("region").String()
	name := d.Get("function_name").String()
	provisionedConcurrentExecutions := d.Get("provisioned_concurrent_executions").Int()

	r := &aws.LambdaProvisionedConcurrencyConfig{
		Address:                         d.Address,
		Region:                          region,
		Name:                            name,
		ProvisionedConcurrentExecutions: provisionedConcurrentExecutions,
	}

	return r
}
