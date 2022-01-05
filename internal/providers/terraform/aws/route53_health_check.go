package aws

import (
	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func getRoute53HealthCheck() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:                "aws_route53_health_check",
		RFunc:               NewRoute53HealthCheck,
		ReferenceAttributes: []string{"alias.0.name"},
	}
}
func NewRoute53HealthCheck(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	r := &aws.Route53HealthCheck{Address: strPtr(d.Address), Type: strPtr(d.Get("type").String()), RequestInterval: strPtr(d.Get("request_interval").String()), MeasureLatency: boolPtr(d.Get("measure_latency").Bool())}
	r.PopulateUsage(u)
	return r.BuildResource()
}
