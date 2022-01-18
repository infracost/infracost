package aws

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"
)

type CloudformationStackSet struct {
	Address                  *string
	TemplateBody             *string
	Region                   *string
	MonthlyHandlerOperations *int64 `infracost_usage:"monthly_handler_operations"`
	MonthlyDurationSecs      *int64 `infracost_usage:"monthly_duration_secs"`
}

var CloudformationStackSetUsageSchema = []*schema.UsageItem{{Key: "monthly_handler_operations", ValueType: schema.Int64, DefaultValue: 0}, {Key: "monthly_duration_secs", ValueType: schema.Int64, DefaultValue: 0}}

func (r *CloudformationStackSet) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *CloudformationStackSet) BuildResource() *schema.Resource {

	if r.TemplateBody != nil && (checkAWS(r.TemplateBody) || checkAlexa(r.TemplateBody) || checkCustom(r.TemplateBody)) {
		return &schema.Resource{
			Name:      *r.Address,
			NoPrice:   true,
			IsSkipped: true, UsageSchema: CloudformationStackSetUsageSchema,
		}
	}

	return &schema.Resource{
		Name:           *r.Address,
		CostComponents: cloudFormationCostComponents(r.Region, r.MonthlyHandlerOperations, r.MonthlyDurationSecs), UsageSchema: CloudformationStackSetUsageSchema,
	}
}
