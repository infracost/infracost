package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func getACMCertificate() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_acm_certificate",
		RFunc: NewAcmCertificate,
	}
}
func NewAcmCertificate(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	r := &aws.AcmCertificate{Address: strPtr(d.Address), Region: strPtr(d.Get("region").String())}
	if !d.IsEmpty("certificate_authority_arn") {
		r.CertificateAuthorityArn = strPtr(d.Get("certificate_authority_arn").String())
	}
	r.PopulateUsage(u)
	return r.BuildResource()
}
