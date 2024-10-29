package aws

import (
	"strings"

	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func getWAFWebACLRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "aws_waf_web_acl",
		CoreRFunc: NewWAFWebACL,
		Notes: []string{
			"Seller fees for Managed Rule Groups from AWS Marketplace are not included. Bot Control is not supported by Terraform.",
		},
	}
}

func NewWAFWebACL(d *schema.ResourceData) schema.CoreResource {
	rules := int64(0)
	ruleGroups := int64(0)

	for _, rule := range d.Get("rules").Array() {
		ruleType := rule.Get("type").String()

		if strings.ToLower(ruleType) == "regular" || strings.ToLower(ruleType) == "rate_based" {
			rules++
		} else if strings.ToLower(ruleType) == "group" {
			ruleGroups++
		}
	}

	r := &aws.WAFWebACL{
		Address:    d.Address,
		Region:     d.Get("region").String(),
		Rules:      rules,
		RuleGroups: ruleGroups,
	}
	return r
}
