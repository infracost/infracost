package aws

import (
	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
)

func GetACMCertificate() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_acm_certificate",
		RFunc: NewACMCertificate,
	}
}

func NewACMCertificate(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	region := d.Get("region").String()

	if d.Get("certificate_authority_arn").Exists() {
		one := decimal.NewFromInt(1)
		return &schema.Resource{
			Name: d.Address,
			CostComponents: []*schema.CostComponent{
				certificateCostComponent(region, "Certificate", "0", &one),
			},
		}
	}

	return &schema.Resource{
		Name:      d.Address,
		NoPrice:   true,
		IsSkipped: true,
	}
}
