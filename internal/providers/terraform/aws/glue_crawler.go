package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func getGlueCrawlerRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_glue_crawler",
		RFunc: newGlueCrawler,
	}
}

func newGlueCrawler(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	region := d.Get("region").String()
	r := &aws.GlueCrawler{
		Address: d.Address,
		Region:  region,
	}
	r.PopulateUsage(u)

	return r.BuildResource()
}
