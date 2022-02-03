package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func getCloudtrailRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_cloudtrail",
		RFunc: newCloudtrail,
	}
}

func newCloudtrail(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	region := d.Get("region").String()
	r := &aws.Cloudtrail{
		Address:                 d.Address,
		Region:                  region,
		IncludeManagementEvents: d.GetBoolOrDefault("include_global_service_events", true),
		IncludeInsightEvents:    len(d.Get("insight_selector").Array()) > 0,
	}
	r.PopulateUsage(u)

	return r.BuildResource()
}
