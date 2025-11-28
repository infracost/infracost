package azure

import (
	"strings"

	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
)

// PrivateDnsResolverOutboundEndpoint struct represents a Azure DNS Private Resolver Outbound Endpoint.
//
// Resource information: https://learn.microsoft.com/en-us/azure/dns/dns-private-resolver-overview
// Pricing information: https://azure.microsoft.com/en-us/pricing/details/dns/
type PrivateDnsResolverOutboundEndpoint struct {
	Address string
	Region  string
}

// CoreType returns the name of this resource type
func (r *PrivateDnsResolverOutboundEndpoint) CoreType() string {
	return "PrivateDnsResolverOutboundEndpoint"
}

// UsageSchema defines a list which represents the usage schema of PrivateDnsResolverOutboundEndpoint.
func (r *PrivateDnsResolverOutboundEndpoint) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{}
}

// PopulateUsage parses the u schema.UsageData into the PrivateDnsResolverOutboundEndpoint.
// It uses the `infracost_usage` struct tags to populate data into the PrivateDnsResolverOutboundEndpoint.
func (r *PrivateDnsResolverOutboundEndpoint) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

// BuildResource builds a schema.Resource from a valid PrivateDnsResolverOutboundEndpoint struct.
// This method is called after the resource is initialised by an IaC provider.
// See providers folder for more information.
func (r *PrivateDnsResolverOutboundEndpoint) BuildResource() *schema.Resource {
	region := r.Region
	switch {
	case strings.HasPrefix(strings.ToLower(region), "usgov"):
		region = "US Gov Zone 1"
	case strings.HasPrefix(strings.ToLower(region), "germany"):
		region = "DE Zone 1"
	case strings.HasPrefix(strings.ToLower(region), "china"):
		region = "Zone 1 (China)"
	default:
		region = "Zone 1"
	}

	return &schema.Resource{
		Name:        r.Address,
		UsageSchema: r.UsageSchema(),
		CostComponents: []*schema.CostComponent{
			{
				Name:            "Outbound endpoint",
				Unit:            "months",
				UnitMultiplier:  decimal.NewFromInt(1),
				MonthlyQuantity: decimalPtr(decimal.NewFromInt(1)),
				ProductFilter: &schema.ProductFilter{
					VendorName:    strPtr("azure"),
					Region:        strPtr(region),
					Service:       strPtr("Azure DNS"),
					ProductFamily: strPtr("Networking"),
					AttributeFilters: []*schema.AttributeFilter{
						{Key: "skuName", ValueRegex: regexPtr("Private Resolver")},
						{Key: "meterName", ValueRegex: regexPtr("Private Resolver Outbound Endpoint")},
					},
				},
			},
		},
	}
}
