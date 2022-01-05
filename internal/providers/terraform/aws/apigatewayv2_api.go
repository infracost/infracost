package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func getAPIGatewayv2APIRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_apigatewayv2_api",
		RFunc: NewAPIgatewayv2API,
	}
}
func NewAPIgatewayv2API(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	r := &aws.APIgatewayv2API{Address: strPtr(d.Address), ProtocolType: strPtr(d.Get("protocol_type").String()), Region: strPtr(d.Get("region").String())}
	r.PopulateUsage(u)
	return r.BuildResource()
}
