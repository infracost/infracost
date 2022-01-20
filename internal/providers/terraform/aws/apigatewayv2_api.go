package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func getAPIGatewayV2APIRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_apigatewayv2_api",
		RFunc: NewAPIGatewayV2API,
	}
}
func NewAPIGatewayV2API(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	r := &aws.APIGatewayV2API{
		Address:      d.Address,
		ProtocolType: d.Get("protocol_type").String(),
		Region:       d.Get("region").String(),
	}

	r.PopulateUsage(u)
	return r.BuildResource()
}
