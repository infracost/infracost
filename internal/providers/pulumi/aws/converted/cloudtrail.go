package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func getCloudtrailRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "aws_cloudtrail",
		RFunc: newCloudtrail,
	}
}

func newCloudtrail(d *schema.ResourceData) schema.CoreResource {
	region := d.Get("region").String()
	r := &aws.Cloudtrail{
		Address:                 d.Address,
		Region:                  region,
		IncludeManagementEvents: d.GetBoolOrDefault("includeGlobalServiceEvents", true),
		IncludeInsightEvents:    len(d.Get("insightSelector").Array()) > 0,
	}

	return r
}
