package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func getWafWebACLRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_waf_web_acl",
		RFunc: NewWafWebACL,
		Notes: []string{
			"Seller fees for Managed Rule Groups from AWS Marketplace are not included. Bot Control is not supported by Terraform.",
		},
	}
}
func NewWafWebACL(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	r := &aws.WafWebACL{Address: strPtr(d.Address), Region: strPtr(d.Get("region").String())}
	if !d.IsEmpty("rules") {
		rulesTypes := make([]string, 0)
		rules := d.Get("rules").Array()
		for _, val := range rules {
			rulesTypes = append(rulesTypes, val.Get("type").String())
		}
		r.RulesTypes = &rulesTypes
	}
	r.PopulateUsage(u)
	return r.BuildResource()
}
