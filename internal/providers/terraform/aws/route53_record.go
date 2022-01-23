package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func getRoute53RecordRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:                "aws_route53_record",
		RFunc:               NewRoute53Record,
		ReferenceAttributes: []string{"alias.0.name"},
	}
}
func NewRoute53Record(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	var aliasType *string

	if len(d.References("alias.0.name")) > 0 && d.References("alias.0.name")[0].Type != "aws_route53_record" {
		aliasType = strPtr("aws_route53_record")
	}

	r := &aws.Route53Record{Address: strPtr(d.Address), AliasType: aliasType}
	r.PopulateUsage(u)
	return r.BuildResource()
}
