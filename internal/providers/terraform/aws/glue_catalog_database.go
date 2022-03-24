package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func getGlueCatalogDatabaseRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_glue_catalog_database",
		RFunc: newGlueCatalogDatabase,
	}
}

func newGlueCatalogDatabase(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	region := d.Get("region").String()
	r := &aws.GlueCatalogDatabase{
		Address: d.Address,
		Region:  region,
	}
	r.PopulateUsage(u)

	return r.BuildResource()
}
