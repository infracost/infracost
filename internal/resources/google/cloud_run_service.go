package google

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"
)

// CloudRunService struct represents <TODO: cloud service short description>.
//
// <TODO: Add any important information about the resource and links to the
// pricing pages or documentation that might be useful to developers in the future, e.g:>
//
// Resource information: https://cloud.google.com/<PATH/TO/RESOURCE>/
// Pricing information: https://cloud.google.com/<PATH/TO/PRICING>/
type CloudRunService struct {
	Address string
	Region  string
}

// CoreType returns the name of this resource type
func (r *CloudRunService) CoreType() string {
	return "CloudRunService"
}

// UsageSchema defines a list which represents the usage schema of CloudRunService.
func (r *CloudRunService) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{
	}
}

// PopulateUsage parses the u schema.UsageData into the CloudRunService.
// It uses the `infracost_usage` struct tags to populate data into the CloudRunService.
func (r *CloudRunService) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

// BuildResource builds a schema.Resource from a valid CloudRunService struct.
// This method is called after the resource is initialised by an IaC provider.
// See providers folder for more information.
func (r *CloudRunService) BuildResource() *schema.Resource {
	costComponents := []*schema.CostComponent{
		// TODO: add cost components
	}

	return &schema.Resource{
		Name:           r.Address,
		UsageSchema:    r.UsageSchema(),
		CostComponents: costComponents,
	}
}

