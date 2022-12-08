package aws

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"
)

// SpotInstanceRequest struct represents <TODO: cloud service short description>.
//
// <TODO: Add any important information about the resource and links to the
// pricing pages or documentation that might be useful to developers in the future, e.g:>
//
// Resource information: https://aws.amazon.com/<PATH/TO/RESOURCE>/
// Pricing information: https://aws.amazon.com/<PATH/TO/PRICING>/
type SpotInstanceRequest struct {
	// "required" arguments
	Address      string
	Region       string
	AMI          string
	InstanceType string
	SpotPrice    string

	// "optional" arguments

	// "usage" arguments
}

var SpotInstanceRequestUsageSchema = []*schema.UsageItem{}

func (r *SpotInstanceRequest) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *SpotInstanceRequest) BuildResource() *schema.Resource {

	instance := &Instance{}

	instance.PurchaseOption = "spot"

  instanceResource := instance.BuildResource()
  instanceResource.UsageSchema = SpotInstanceRequestUsageSchema

	instanceResource.CostComponents = []*schema.CostComponent{
		// TODO: add cost components
	}

  return instanceResource

	//return &schema.Resource{
	//	Name:           instance.Address,
	//	UsageSchema:    SpotInstanceRequestUsageSchema,
	//	CostComponents: costComponents,
	//}
}
