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
	isAlias := false

	aliasRefs := d.References("alias.0.name")
	if len(aliasRefs) > 0 && aliasRefs[0].Type != "aws_route53_record" {
		isAlias = true
	}

	r := &aws.Route53Record{
		Address: d.Address,
		IsAlias: isAlias,
	}

	r.PopulateUsage(u)
	return r.BuildResource()
}
