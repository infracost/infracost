package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func GetAPIGatewayv2ApiRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_apigatewayv2_api",
		RFunc: NewAPIGatewayv2Api,
	}
}
func NewAPIGatewayv2Api(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	r := &aws.APIGatewayv2Api{Address: strPtr(d.Address), ProtocolType: strPtr(d.Get("protocol_type").String()), Region: strPtr(d.Get("region").String())}
	r.PopulateUsage(u)
	return r.BuildResource()
}
