package azure

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"
)

// SentinelSentinelDataConnectorOffice365 struct represents <TODO: cloud service short description>.
//
// <TODO: Add any important information about the resource and links to the
// pricing pages or documentation that might be useful to developers in the future, e.g:>
//
// Resource information: https://azure.microsoft.com/<PATH/TO/RESOURCE>/
// Pricing information: https://azure.microsoft.com/<PATH/TO/PRICING>/
type SentinelSentinelDataConnectorOffice365 struct {
	Address string
	Region  string
}

// SentinelSentinelDataConnectorOffice365UsageSchema defines a list which represents the usage schema of SentinelSentinelDataConnectorOffice365.
var SentinelSentinelDataConnectorOffice365UsageSchema = []*schema.UsageItem{}

// PopulateUsage parses the u schema.UsageData into the SentinelSentinelDataConnectorOffice365.
// It uses the `infracost_usage` struct tags to populate data into the SentinelSentinelDataConnectorOffice365.
func (r *SentinelSentinelDataConnectorOffice365) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

// BuildResource builds a schema.Resource from a valid SentinelSentinelDataConnectorOffice365 struct.
// This method is called after the resource is initialised by an IaC provider.
// See providers folder for more information.
func (r *SentinelSentinelDataConnectorOffice365) BuildResource() *schema.Resource {
	return &schema.Resource{
		Name:        r.Address,
		IsSkipped:   true,
		NoPrice:     true,
		UsageSchema: SentinelSentinelDataConnectorOffice365UsageSchema,
	}
}
