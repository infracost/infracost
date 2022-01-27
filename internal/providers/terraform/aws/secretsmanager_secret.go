package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func getSecretsManagerSecret() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_secretsmanager_secret",
		RFunc: NewSecretsManagerSecret,
	}
}

func NewSecretsManagerSecret(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	r := &aws.SecretsManagerSecret{
		Address: d.Address,
		Region:  d.Get("region").String(),
	}

	r.PopulateUsage(u)
	return r.BuildResource()
}
