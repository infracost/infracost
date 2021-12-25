package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
	"github.com/tidwall/gjson"
)

func GetACMCertificate() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_acm_certificate",
		RFunc: NewACMCertificate,
	}
}
func NewACMCertificate(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	r := &aws.ACMCertificate{Address: strPtr(d.Address), Region: strPtr(d.Get("region").String())}
	if d.Get("certificate_authority_arn").Exists() && d.Get("certificate_authority_arn").Type != gjson.Null {
		r.CertificateAuthorityArn = strPtr(d.Get("certificate_authority_arn").String())
	}
	r.PopulateUsage(u)
	return r.BuildResource()
}
