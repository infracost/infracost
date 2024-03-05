package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func getRoute53RecordRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:                "aws_route53_record",
		CoreRFunc:           NewRoute53Record,
		ReferenceAttributes: []string{"alias.0.name"},
	}
}
func NewRoute53Record(d *schema.ResourceData) schema.CoreResource {
	isAlias := false

	aliasRefs := d.References("alias.0.name")
	if len(aliasRefs) > 0 && aliasRefs[0].Type != "aws_route53_record" {
		isAlias = true
	}

	r := &aws.Route53Record{
		Address: d.Address,
		IsAlias: isAlias,
	}
	return r
}
