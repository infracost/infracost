package google

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"

	"github.com/shopspring/decimal"
)

type LoggingFolderBucketConfig struct {
	Address              *string
	MonthlyLoggingDataGb *float64 `infracost_usage:"monthly_logging_data_gb"`
}

var LoggingFolderBucketConfigUsageSchema = []*schema.UsageItem{{Key: "monthly_logging_data_gb", ValueType: schema.Float64, DefaultValue: 0}}

func (r *LoggingFolderBucketConfig) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *LoggingFolderBucketConfig) BuildResource() *schema.Resource {
	var loggingData *decimal.Decimal
	if r.MonthlyLoggingDataGb != nil {
		loggingData = decimalPtr(decimal.NewFromFloat(*r.MonthlyLoggingDataGb))
	}

	return &schema.Resource{
		Name:           *r.Address,
		CostComponents: loggingCostComponent(loggingData), UsageSchema: LoggingFolderBucketConfigUsageSchema,
	}
}
