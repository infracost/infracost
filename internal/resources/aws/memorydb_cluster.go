package aws

import (
	"fmt"
	"strings"
	"github.com/infracost/infracost/internal/logging"
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
)

// MemoryDBCluster represents an AWS MemoryDB cluster
//
// Resource information: https://aws.amazon.com/memorydb/
// Pricing information: https://aws.amazon.com/memorydb/pricing/
//
// Pricing notes:
// - MemoryDB uses the same pricing structure as ElastiCache
// - Valkey engine is 30% cheaper than Redis OSS
// - Data written for Valkey is free up to 10TB/month, then $0.04/GB
// - Data written for Redis OSS is $0.20/GB for all data
// - Snapshot storage is $0.021/GB-month beyond the first day of retention
type MemoryDBCluster struct {
	Address                       string
	Region                        string
	NodeType                      string
	Engine                        string // "redis" or "valkey"
	NumShards                     int64
	ReplicasPerShard              int64
	SnapshotRetentionLimit        int64

	// Usage parameters
	MonthlyDataWrittenGB          *float64 `infracost_usage:"monthly_data_written_gb"`
	SnapshotStorageSizeGB         *float64 `infracost_usage:"snapshot_storage_size_gb"`
	ReservedInstanceTerm          *string  `infracost_usage:"reserved_instance_term"`
	ReservedInstancePaymentOption *string  `infracost_usage:"reserved_instance_payment_option"`

	// Autoscaling support
	AppAutoscalingTarget []*AppAutoscalingTarget
}

func (r *MemoryDBCluster) CoreType() string {
	return "MemoryDBCluster"
}

func (r *MemoryDBCluster) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{
		{Key: "monthly_data_written_gb", ValueType: schema.Float64, DefaultValue: 0},
		{Key: "snapshot_storage_size_gb", ValueType: schema.Float64, DefaultValue: 0},
		{Key: "reserved_instance_term", DefaultValue: "", ValueType: schema.String},
		{Key: "reserved_instance_payment_option", DefaultValue: "", ValueType: schema.String},
	}
}

func (r *MemoryDBCluster) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *MemoryDBCluster) BuildResource() *schema.Resource {
	engine := r.Engine
	if engine == "" {
		// Default engine is Redis OSS if not specified
		engine = "redis"
	}

	// Check for autoscaling configuration
	var autoscaling bool
	shards := r.NumShards
	replicasPerShard := r.ReplicasPerShard

	for _, target := range r.AppAutoscalingTarget {
		switch target.ScalableDimension {
		case "memorydb:cluster:Shards":
			autoscaling = true
			if target.Capacity != nil {
				shards = *target.Capacity
			} else {
				shards = target.MinCapacity
			}
		case "memorydb:cluster:ReplicasPerShard":
			autoscaling = true
			if target.Capacity != nil {
				replicasPerShard = *target.Capacity
			} else {
				replicasPerShard = target.MinCapacity
			}
		}
	}

	// Calculate total number of nodes
	totalNodes := shards * (replicasPerShard + 1) // +1 for primary node in each shard

	costComponents := []*schema.CostComponent{
		r.memoryDBCostComponent(totalNodes, autoscaling),
	}

	// Add data written cost component
	costComponents = append(costComponents, r.dataWrittenCostComponent())

	// Add snapshot storage cost component if retention is more than 1 day
	if r.SnapshotRetentionLimit > 1 {
		costComponents = append(costComponents, r.backupStorageCostComponent())
	}

	return &schema.Resource{
		Name:           r.Address,
		UsageSchema:    r.UsageSchema(),
		CostComponents: costComponents,
	}
}

func (r *MemoryDBCluster) memoryDBCostComponent(totalNodes int64, autoscaling bool) *schema.CostComponent {
	purchaseOptionLabel := "on-demand"
	priceFilter := &schema.PriceFilter{
		PurchaseOption: strPtr("on_demand"),
	}

	// Handle reserved instances
	if r.ReservedInstanceTerm != nil {
		resolver := &elasticacheReservationResolver{
			term:          strVal(r.ReservedInstanceTerm),
			paymentOption: strVal(r.ReservedInstancePaymentOption),
			cacheNodeType: r.NodeType,
		}
		reservedFilter, err := resolver.PriceFilter()
		if err != nil {
			logging.Logger.Warn().Msg(err.Error())
		} else {
			priceFilter = reservedFilter
		}
		purchaseOptionLabel = "reserved"
	}

	// Build the name with appropriate parameters
	nameParams := []string{purchaseOptionLabel, r.NodeType}
	if autoscaling {
		nameParams = append(nameParams, "autoscaling")
	}

	// Since MemoryDB pricing isn't directly available in the AWS pricing API,
	// we'll use ElastiCache Redis pricing as it's the same

	// Create a custom price component with a name that reflects MemoryDB
	var name string
	if strings.ToLower(r.Engine) == "valkey" {
		name = fmt.Sprintf("MemoryDB instance (valkey, %s)", strings.Join(nameParams, ", "))
	} else {
		name = fmt.Sprintf("MemoryDB instance (%s)", strings.Join(nameParams, ", "))
	}

	// Create a cost component with a custom price based on the instance type
	// MemoryDB uses the same pricing as ElastiCache Redis
	costComponent := &schema.CostComponent{
		Name:           name,
		Unit:           "hours",
		UnitMultiplier: decimal.NewFromInt(1),
		HourlyQuantity: decimalPtr(decimal.NewFromInt(totalNodes)),
		UsageBased:     autoscaling,
	}

	// Set the price based on the instance type
	// These prices are for us-east-1, other regions may vary
	price := getMemoryDBInstancePrice(r.NodeType)
	costComponent.SetCustomPrice(decimalPtr(decimal.NewFromFloat(price)))

	// Apply a 30% discount for Valkey
	if strings.ToLower(r.Engine) == "valkey" {
		// For Valkey, we'll use Redis pricing but apply a 30% discount
		costComponent.MonthlyDiscountPerc = 0.3 // 30% discount
	}

	return costComponent
}

func (r *MemoryDBCluster) dataWrittenCostComponent() *schema.CostComponent {
	var monthlyDataWrittenGB *decimal.Decimal

	if r.MonthlyDataWrittenGB != nil {
		monthlyDataWrittenGB = decimalPtr(decimal.NewFromFloat(*r.MonthlyDataWrittenGB))
	}

	// For Valkey, data written is free up to 10TB/month
	if strings.ToLower(r.Engine) == "valkey" {
		var dataWrittenGB float64 = 0
		if r.MonthlyDataWrittenGB != nil {
			dataWrittenGB = *r.MonthlyDataWrittenGB
		}

		// Calculate the billable amount (over 10TB)
		var billableGB float64 = 0
		if dataWrittenGB > 10240 { // 10TB in GB
			billableGB = dataWrittenGB - 10240
		}

		// If there's no billable data, return a component showing it's free
		if billableGB <= 0 {
			// For free tier, we'll use a custom price of 0
			costComponent := &schema.CostComponent{
				Name:            "Data written (free up to 10TB)",
				Unit:            "GB",
				UnitMultiplier:  decimal.NewFromInt(1),
				MonthlyQuantity: monthlyDataWrittenGB,
				UsageBased:      true,
			}
			costComponent.SetCustomPrice(decimalPtr(decimal.Zero))
			return costComponent
		}

		// Otherwise, return a component for the billable amount
		// For Valkey over 10TB, we'll use a custom price of $0.04/GB
		costComponent := &schema.CostComponent{
			Name:            "Data written (over 10TB)",
			Unit:            "GB",
			UnitMultiplier:  decimal.NewFromInt(1),
			MonthlyQuantity: decimalPtr(decimal.NewFromFloat(billableGB)),
			UsageBased:      true,
		}
		costComponent.SetCustomPrice(decimalPtr(decimal.NewFromFloat(0.04)))
		return costComponent
	}

	// For Redis, all data written is charged at $0.20/GB
	costComponent := &schema.CostComponent{
		Name:            "Data written",
		Unit:            "GB",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: monthlyDataWrittenGB,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(r.Region),
			Service:       strPtr("AmazonElastiCache"),
			ProductFamily: strPtr("Cache Instance"),
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr("on_demand"),
		},
		UsageBased: true,
	}
	// Since the pricing API doesn't have data written pricing, we'll use a custom price
	costComponent.SetCustomPrice(decimalPtr(decimal.NewFromFloat(0.20)))
	return costComponent
}

func (r *MemoryDBCluster) backupStorageCostComponent() *schema.CostComponent {
	var monthlyBackupStorageGB *decimal.Decimal

	backupRetention := r.SnapshotRetentionLimit - 1

	if r.SnapshotStorageSizeGB != nil {
		snapshotSize := decimal.NewFromFloat(*r.SnapshotStorageSizeGB)
		monthlyBackupStorageGB = decimalPtr(snapshotSize.Mul(decimal.NewFromInt(backupRetention)))
	}

	costComponent := &schema.CostComponent{
		Name:            "Backup storage",
		Unit:            "GB",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: monthlyBackupStorageGB,
		UsageBased:      true,
	}

	// Set a custom price for backup storage
	costComponent.SetCustomPrice(decimalPtr(decimal.NewFromFloat(0.085)))

	return costComponent
}

// getMemoryDBInstancePrice returns the hourly price for a given MemoryDB instance type
// These prices are for us-east-1 and are the same as ElastiCache Redis
// Prices from: https://aws.amazon.com/memorydb/pricing/
func getMemoryDBInstancePrice(instanceType string) float64 {
	prices := map[string]float64{
		// T4G instances
		"db.t4g.small":   0.068,
		"db.t4g.medium":  0.136,

		// R6G instances
		"db.r6g.large":    0.40,
		"db.r6g.xlarge":   0.80,
		"db.r6g.2xlarge":  1.60,
		"db.r6g.4xlarge":  3.20,
		"db.r6g.8xlarge":  6.40,
		"db.r6g.12xlarge": 9.60,
		"db.r6g.16xlarge": 12.80,

		// R6GD instances
		"db.r6gd.xlarge":   0.70,
		"db.r6gd.2xlarge":  1.40,
		"db.r6gd.4xlarge":  2.80,
		"db.r6gd.8xlarge":  5.60,

		// M6G instances
		"db.m6g.large":    0.20,
		"db.m6g.xlarge":   0.40,
		"db.m6g.2xlarge":  0.80,
		"db.m6g.4xlarge":  1.60,
		"db.m6g.8xlarge":  3.20,
		"db.m6g.12xlarge": 4.80,
		"db.m6g.16xlarge": 6.40,
	}

	if price, ok := prices[instanceType]; ok {
		return price
	}

	// Default price if instance type is not found
	// This is a fallback to avoid "not found" errors
	logging.Logger.Warn().Msgf("Price not found for MemoryDB instance type: %s, using default price", instanceType)
	return 0.40 // Default to r6g.large price
}


