package aws

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"
)

type CloudFormationStackSet struct {
	Address                  string
	Region                   string
	TemplateBody             string
	MonthlyHandlerOperations *int64 `infracost_usage:"monthly_handler_operations"`
	MonthlyDurationSecs      *int64 `infracost_usage:"monthly_duration_secs"`
}

func (r *CloudFormationStackSet) CoreType() string {
	return "CloudFormationStackSet"
}

func (r *CloudFormationStackSet) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{
		{Key: "monthly_handler_operations", ValueType: schema.Int64, DefaultValue: 0},
		{Key: "monthly_duration_secs", ValueType: schema.Int64, DefaultValue: 0},
	}
}

func (r *CloudFormationStackSet) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *CloudFormationStackSet) BuildResource() *schema.Resource {
	stack := &CloudFormationStack{
		Region:                   r.Region,
		TemplateBody:             r.TemplateBody,
		MonthlyHandlerOperations: r.MonthlyHandlerOperations,
		MonthlyDurationSecs:      r.MonthlyDurationSecs,
	}

	if stack.checkAWS() || stack.checkAlexa() || stack.checkCustom() {
		return &schema.Resource{
			Name:        r.Address,
			NoPrice:     true,
			IsSkipped:   true,
			UsageSchema: r.UsageSchema(),
		}
	}

	return &schema.Resource{
		Name:           r.Address,
		CostComponents: stack.costComponents(),
		UsageSchema:    r.UsageSchema(),
	}
}
