package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func getSecretsManagerSecret() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_secretsmanager_secret",
		RFunc: NewSecretsmanagerSecret,
	}
}
func NewSecretsmanagerSecret(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	r := &aws.SecretsmanagerSecret{Address: strPtr(d.Address), Region: strPtr(d.Get("region").String())}
	r.PopulateUsage(u)
	return r.BuildResource()
}
