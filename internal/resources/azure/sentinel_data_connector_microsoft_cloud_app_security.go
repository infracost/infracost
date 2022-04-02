package azure

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"
)

// SentinelDataConnectorMicrosoftCloudAppSecurity struct represents <TODO: cloud service short description>.
//
// <TODO: Add any important information about the resource and links to the
// pricing pages or documentation that might be useful to developers in the future, e.g:>
//
// Resource information: https://azure.microsoft.com/<PATH/TO/RESOURCE>/
// Pricing information: https://azure.microsoft.com/<PATH/TO/PRICING>/
type SentinelDataConnectorMicrosoftCloudAppSecurity struct {
	Address string
	Region  string
}

// SentinelDataConnectorMicrosoftCloudAppSecurityUsageSchema defines a list which represents the usage schema of SentinelDataConnectorMicrosoftCloudAppSecurity.
var SentinelDataConnectorMicrosoftCloudAppSecurityUsageSchema = []*schema.UsageItem{}

// PopulateUsage parses the u schema.UsageData into the SentinelDataConnectorMicrosoftCloudAppSecurity.
// It uses the `infracost_usage` struct tags to populate data into the SentinelDataConnectorMicrosoftCloudAppSecurity.
func (r *SentinelDataConnectorMicrosoftCloudAppSecurity) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

// BuildResource builds a schema.Resource from a valid SentinelDataConnectorMicrosoftCloudAppSecurity struct.
// This method is called after the resource is initialised by an IaC provider.
// See providers folder for more information.
func (r *SentinelDataConnectorMicrosoftCloudAppSecurity) BuildResource() *schema.Resource {
	return &schema.Resource{
		Name:        r.Address,
		IsSkipped:   true,
		NoPrice:     true,
		UsageSchema: SentinelDataConnectorMicrosoftCloudAppSecurityUsageSchema,
	}
}
