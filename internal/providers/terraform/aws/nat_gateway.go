package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func GetNATGatewayRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_nat_gateway",
		RFunc: NewNATGateway,
	}
}

func NewNATGateway(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	region := d.Get("region").String()

	// We initialize the args and populate the defaults and usage data.
	args := &aws.NATGatewayArguments{}
	args.PopulateArgs(u)

	// Then we can override the fields with IAC fields.
	// We can have a utility function to print an info log
	// that users could know the following fields were overridden.
	args.Region = &region
	args.Address = &d.Address

	return aws.NewNATGateway(args)
}
