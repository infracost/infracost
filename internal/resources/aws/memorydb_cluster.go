package aws

import (
	"fmt"
	"strings"

	"github.com/infracost/infracost/internal/logging"
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
)

// Constants for MemoryDB pricing and configuration
// Last price update: May 2023
const (
	// Engine types
	MemoryDBEngineRedis  = "redis"
	MemoryDBEngineValkey = "valkey"

	// Pricing constants
	ValkeyCostDiscount          = 0.3   // 30% discount for Valkey engine
	ValkeyCostFreeDataWrittenGB = 10240 // 10TB in GB
	RedisDataWrittenCostPerGB   = 0.20  // $0.20/GB for Redis
	ValkeyDataWrittenCostPerGB  = 0.04  // $0.04/GB for Valkey (over free tier)
	BackupStorageCostPerGB      = 0.085 // $0.085/GB for backup storage

	// Default instance type for fallback
	DefaultInstanceType = "db.r6g.large"
)

// MemoryDBCluster represents an AWS MemoryDB cluster
//
// Resource information: https://aws.amazon.com/memorydb/
// Pricing information: https://aws.amazon.com/memorydb/pricing/
//
// Pricing notes:
// - MemoryDB has its own pricing, which is different from ElastiCache
// - Valkey engine is 30% cheaper than Redis OSS
// - Data written for Valkey is free up to 10TB/month, then $0.04/GB
// - Data written for Redis OSS is $0.20/GB for all data
// - Snapshot storage is $0.085/GB-month beyond the first day of retention
type MemoryDBCluster struct {
	Address                string
	Region                 string
	NodeType               string
	Engine                 string // "redis" or "valkey"
	NumShards              int64
	ReplicasPerShard       int64
	SnapshotRetentionLimit int64

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
	// Validate and set default engine
	engine := strings.ToLower(r.Engine)
	if engine == "" {
		// Default engine is Redis OSS if not specified
		engine = MemoryDBEngineRedis
	} else if engine != MemoryDBEngineRedis && engine != MemoryDBEngineValkey {
		// Log warning for unknown engine type and default to Redis
		logging.Logger.Warn().Msgf("Unknown engine type %s for MemoryDB cluster %s, defaulting to redis", r.Engine, r.Address)
		engine = MemoryDBEngineRedis
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

	// Handle reserved instances with validation
	if r.ReservedInstanceTerm != nil {
		// Validate reserved instance term
		validTerms := []string{"1_year", "3_year"}
		term := strVal(r.ReservedInstanceTerm)

		if !stringInSlice(validTerms, term) {
			logging.Logger.Warn().Msgf("Invalid reserved_instance_term for MemoryDB cluster %s. Expected: %s. Got: %s. Using on-demand pricing.",
				r.Address, strings.Join(validTerms, ", "), term)
		} else if r.ReservedInstancePaymentOption == nil {
			logging.Logger.Warn().Msgf("Reserved instance term specified without payment option for %s. Using on-demand pricing.", r.Address)
		} else {
			// Validate payment option
			validOptions := []string{"all_upfront", "partial_upfront", "no_upfront"}
			paymentOption := strVal(r.ReservedInstancePaymentOption)

			if !stringInSlice(validOptions, paymentOption) {
				logging.Logger.Warn().Msgf("Invalid reserved_instance_payment_option for MemoryDB cluster %s. Expected: %s. Got: %s. Using on-demand pricing.",
					r.Address, strings.Join(validOptions, ", "), paymentOption)
			} else {
				purchaseOptionLabel = "reserved"
			}
		}
	}

	// Build the name with appropriate parameters
	nameParams := []string{purchaseOptionLabel, r.NodeType}
	if autoscaling {
		nameParams = append(nameParams, "autoscaling")
	}

	// Create a custom price component with a name that reflects MemoryDB
	var name string
	if strings.ToLower(r.Engine) == MemoryDBEngineValkey {
		name = fmt.Sprintf("MemoryDB instance (valkey, %s)", strings.Join(nameParams, ", "))
	} else {
		name = fmt.Sprintf("MemoryDB instance (%s)", strings.Join(nameParams, ", "))
	}

	// Create a cost component with a custom price based on the instance type
	costComponent := &schema.CostComponent{
		Name:           name,
		Unit:           "hours",
		UnitMultiplier: decimal.NewFromInt(1),
		HourlyQuantity: decimalPtr(decimal.NewFromInt(totalNodes)),
		UsageBased:     autoscaling,
	}

	// Get base price for the instance type with region consideration
	price := getMemoryDBInstancePrice(r.NodeType, r.Region)

	// Apply reserved instance discount if applicable
	if purchaseOptionLabel == "reserved" {
		price = applyReservedInstanceDiscount(price, strVal(r.ReservedInstanceTerm), strVal(r.ReservedInstancePaymentOption))
	}

	costComponent.SetCustomPrice(decimalPtr(decimal.NewFromFloat(price)))

	// Apply a 30% discount for Valkey
	if strings.ToLower(r.Engine) == MemoryDBEngineValkey {
		costComponent.MonthlyDiscountPerc = ValkeyCostDiscount
	}

	return costComponent
}

func (r *MemoryDBCluster) dataWrittenCostComponent() *schema.CostComponent {
	var monthlyDataWrittenGB *decimal.Decimal

	if r.MonthlyDataWrittenGB != nil {
		monthlyDataWrittenGB = decimalPtr(decimal.NewFromFloat(*r.MonthlyDataWrittenGB))
	}

	// Handle data written pricing based on engine type
	engine := strings.ToLower(r.Engine)

	if engine == MemoryDBEngineValkey {
		return r.valkeyDataWrittenCostComponent(monthlyDataWrittenGB)
	}

	// Default to Redis pricing
	return r.redisDataWrittenCostComponent(monthlyDataWrittenGB)
}

// Helper method for Valkey data written pricing
func (r *MemoryDBCluster) valkeyDataWrittenCostComponent(monthlyDataWrittenGB *decimal.Decimal) *schema.CostComponent {
	var dataWrittenGB float64 = 0
	if r.MonthlyDataWrittenGB != nil {
		dataWrittenGB = *r.MonthlyDataWrittenGB
	}

	// Calculate the billable amount (over 10TB)
	var billableGB float64 = 0
	if dataWrittenGB > ValkeyCostFreeDataWrittenGB {
		billableGB = dataWrittenGB - ValkeyCostFreeDataWrittenGB
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
	costComponent := &schema.CostComponent{
		Name:            "Data written (over 10TB)",
		Unit:            "GB",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: decimalPtr(decimal.NewFromFloat(billableGB)),
		UsageBased:      true,
	}
	costComponent.SetCustomPrice(decimalPtr(decimal.NewFromFloat(ValkeyDataWrittenCostPerGB)))
	return costComponent
}

// Helper method for Redis data written pricing
func (r *MemoryDBCluster) redisDataWrittenCostComponent(monthlyDataWrittenGB *decimal.Decimal) *schema.CostComponent {
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
	costComponent.SetCustomPrice(decimalPtr(decimal.NewFromFloat(RedisDataWrittenCostPerGB)))
	return costComponent
}

func (r *MemoryDBCluster) backupStorageCostComponent() *schema.CostComponent {
	var monthlyBackupStorageGB *decimal.Decimal

	// Calculate backup retention (days beyond the first day)
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
	costComponent.SetCustomPrice(decimalPtr(decimal.NewFromFloat(BackupStorageCostPerGB)))

	return costComponent
}

// getMemoryDBInstancePrice returns the hourly price for a given MemoryDB instance type and region
// Base prices are for us-east-1 and are from the AWS MemoryDB pricing page
// Prices from: https://aws.amazon.com/memorydb/pricing/
// Last price update: May 2023
func getMemoryDBInstancePrice(instanceType string, region string) float64 {
	// Base prices for us-east-1
	prices := map[string]float64{
		// T4G instances
		"db.t4g.small":  0.038,
		"db.t4g.medium": 0.076,

		// T3 instances
		"db.t3.small":  0.047,
		"db.t3.medium": 0.094,

		// R6G instances
		"db.r6g.large":    0.228,
		"db.r6g.xlarge":   0.456,
		"db.r6g.2xlarge":  0.912,
		"db.r6g.4xlarge":  1.824,
		"db.r6g.8xlarge":  3.648,
		"db.r6g.12xlarge": 5.472,
		"db.r6g.16xlarge": 7.296,

		// R6GD instances
		"db.r6gd.xlarge":  0.399,
		"db.r6gd.2xlarge": 0.798,
		"db.r6gd.4xlarge": 1.596,
		"db.r6gd.8xlarge": 3.192,

		// M6G instances
		"db.m6g.large":    0.114,
		"db.m6g.xlarge":   0.228,
		"db.m6g.2xlarge":  0.456,
		"db.m6g.4xlarge":  0.912,
		"db.m6g.8xlarge":  1.824,
		"db.m6g.12xlarge": 2.736,
		"db.m6g.16xlarge": 3.648,
	}

	// Get base price for the instance type
	var basePrice float64
	var ok bool
	if basePrice, ok = prices[instanceType]; !ok {
		// Default price if instance type is not found
		logging.Logger.Warn().Msgf("Price not found for MemoryDB instance type: %s, using default price for %s", instanceType, DefaultInstanceType)
		basePrice = prices[DefaultInstanceType]
	}

	// Apply regional price adjustment
	return basePrice * getRegionPriceMultiplier(region)
}

// getRegionPriceMultiplier returns a multiplier for prices in different AWS regions
// These multipliers are approximations based on typical regional price differences
func getRegionPriceMultiplier(region string) float64 {
	// Regional price multipliers (approximate)
	regionMultipliers := map[string]float64{
		// US Regions
		"us-east-1": 1.00, // N. Virginia (base)
		"us-east-2": 1.00, // Ohio
		"us-west-1": 1.06, // N. California
		"us-west-2": 1.00, // Oregon

		// Canada
		"ca-central-1": 1.08, // Canada

		// South America
		"sa-east-1": 1.25, // São Paulo

		// Europe
		"eu-central-1": 1.09, // Frankfurt
		"eu-west-1":    1.03, // Ireland
		"eu-west-2":    1.07, // London
		"eu-west-3":    1.09, // Paris
		"eu-north-1":   1.07, // Stockholm
		"eu-south-1":   1.09, // Milan

		// Asia Pacific
		"ap-east-1":      1.20, // Hong Kong
		"ap-south-1":     1.12, // Mumbai
		"ap-northeast-1": 1.15, // Tokyo
		"ap-northeast-2": 1.15, // Seoul
		"ap-northeast-3": 1.15, // Osaka
		"ap-southeast-1": 1.15, // Singapore
		"ap-southeast-2": 1.15, // Sydney

		// Middle East
		"me-south-1": 1.15, // Bahrain

		// Africa
		"af-south-1": 1.15, // Cape Town
	}

	if multiplier, ok := regionMultipliers[region]; ok {
		return multiplier
	}

	// Default to base price if region not found
	logging.Logger.Warn().Msgf("Price multiplier not found for region: %s, using base price", region)
	return 1.0
}

// applyReservedInstanceDiscount applies the appropriate discount for reserved instances
// based on the term and payment option
func applyReservedInstanceDiscount(basePrice float64, term string, paymentOption string) float64 {
	// Discount percentages for different reservation options
	// These are approximate and based on typical RI discounts
	discounts := map[string]map[string]float64{
		"1_year": {
			"no_upfront":      0.20, // 20% discount
			"partial_upfront": 0.30, // 30% discount
			"all_upfront":     0.35, // 35% discount
		},
		"3_year": {
			"no_upfront":      0.40, // 40% discount
			"partial_upfront": 0.50, // 50% discount
			"all_upfront":     0.60, // 60% discount
		},
	}

	if termDiscounts, ok := discounts[term]; ok {
		if discount, ok := termDiscounts[paymentOption]; ok {
			return basePrice * (1 - discount)
		}
	}

	// If no valid discount found, return the base price
	logging.Logger.Warn().Msgf("No valid discount found for term: %s, payment option: %s. Using on-demand price.", term, paymentOption)
	return basePrice
}
