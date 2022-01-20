package aws

import (
	"fmt"

	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"

	"strings"

	"github.com/shopspring/decimal"
)

type CodeBuildProject struct {
	Address          string
	Region           string
	ComputeType      string
	EnvironmentType  string
	MonthlyBuildMins *int64 `infracost_usage:"monthly_build_mins"`
}

var CodeBuildProjectUsageSchema = []*schema.UsageItem{
	{Key: "monthly_build_mins", ValueType: schema.Int64, DefaultValue: 0},
}

func (r *CodeBuildProject) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *CodeBuildProject) BuildResource() *schema.Resource {
	var monthlyBuildMinutes *decimal.Decimal
	if r.MonthlyBuildMins != nil {
		monthlyBuildMinutes = decimalPtr(decimal.NewFromInt(*r.MonthlyBuildMins))
	}

	return &schema.Resource{
		Name: r.Address,
		CostComponents: []*schema.CostComponent{
			{
				Name:            r.nameLabel(),
				Unit:            "minutes",
				UnitMultiplier:  decimal.NewFromInt(1),
				MonthlyQuantity: monthlyBuildMinutes,
				ProductFilter: &schema.ProductFilter{
					VendorName:    strPtr("aws"),
					Region:        strPtr(r.Region),
					Service:       strPtr("CodeBuild"),
					ProductFamily: strPtr("Compute"),
					AttributeFilters: []*schema.AttributeFilter{
						{Key: "usagetype", ValueRegex: strPtr(fmt.Sprintf("/%s:%s/", r.mapEnvironmentType(), r.mapComputeType()))},
					},
				},
			},
		},
		UsageSchema: CodeBuildProjectUsageSchema,
	}
}

func (r *CodeBuildProject) nameLabel() string {
	name := ""
	switch r.EnvironmentType {
	case "WINDOWS_SERVER_2019_CONTAINER":
		name = "Windows ("
		name += strings.Replace(strings.ToLower(strings.SplitAfter(r.ComputeType, "BUILD_")[1]), "_", ".", 1) + ")"
	case "ARM_CONTAINER":
		name = "Linux (arm1.large)"
	case "LINUX_GPU_CONTAINER":
		name = "Linux (gpu1.large)"
	default:
		name = "Linux ("
		name += strings.Replace(strings.ToLower(strings.SplitAfter(r.ComputeType, "BUILD_")[1]), "_", ".", 1) + ")"
	}

	return name
}

func (r *CodeBuildProject) mapEnvironmentType() string {
	switch r.EnvironmentType {
	case "LINUX_CONTAINER":
		return "Linux"
	case "LINUX_GPU_CONTAINER":
		return "LinuxGPU"
	case "ARM_CONTAINER":
		return "ARM"
	case "WINDOWS_SERVER_2019_CONTAINER":
		return "Windows"
	default:
		return ""
	}
}

func (r *CodeBuildProject) mapComputeType() string {
	switch r.ComputeType {
	case "BUILD_GENERAL1_SMALL":
		return "g1.small"
	case "BUILD_GENERAL1_MEDIUM":
		return "g1.medium"
	case "BUILD_GENERAL1_LARGE":
		return "g1.large"
	case "BUILD_GENERAL1_2XLARGE":
		return "g1.2xlarge"
	default:
		return ""
	}
}
