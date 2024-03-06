package azure

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"
)

type DNSPtrRecord struct {
	Address        string
	Region         string
	MonthlyQueries *int64 `infracost_usage:"monthly_queries"`
}

func (r *DNSPtrRecord) CoreType() string {
	return "DNSPtrRecord"
}

func (r *DNSPtrRecord) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{{Key: "monthly_queries", ValueType: schema.Int64, DefaultValue: 0}}
}

func (r *DNSPtrRecord) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *DNSPtrRecord) BuildResource() *schema.Resource {
	return &schema.Resource{
		Name:           r.Address,
		CostComponents: dnsQueriesCostComponent(r.Region, r.MonthlyQueries),
		UsageSchema:    r.UsageSchema(),
	}
}
