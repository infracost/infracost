package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func getRoute53HealthCheck() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_route53_health_check",
		RFunc: NewRoute53HealthCheck,
	}
}

func NewRoute53HealthCheck(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	r := &aws.Route53HealthCheck{
		Address:         d.Address,
		Type:            d.Get("type").String(),
		RequestInterval: d.Get("request_interval").String(),
		MeasureLatency:  d.Get("measure_latency").Bool(),
	}
	r.PopulateUsage(u)
	return r.BuildResource()
}
