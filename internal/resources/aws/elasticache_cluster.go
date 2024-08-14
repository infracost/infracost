package aws

import (
	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"github.com/infracost/infracost/internal/logging"
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"

	"fmt"
	"strings"

	"github.com/shopspring/decimal"
)

type ElastiCacheCluster struct {
	Address                       string
	Region                        string
	HasReplicationGroup           bool
	NodeType                      string
	Engine                        string
	CacheNodes                    int64
	SnapshotRetentionLimit        int64
	SnapshotStorageSizeGB         *float64 `infracost_usage:"snapshot_storage_size_gb"`
	ReservedInstanceTerm          *string  `infracost_usage:"reserved_instance_term"`
	ReservedInstancePaymentOption *string  `infracost_usage:"reserved_instance_payment_option"`
}

func (r *ElastiCacheCluster) CoreType() string {
	return "ElastiCacheCluster"
}

func (r *ElastiCacheCluster) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{
		{Key: "snapshot_storage_size_gb", ValueType: schema.Float64, DefaultValue: 0},
		{Key: "reserved_instance_term", DefaultValue: "", ValueType: schema.String},
		{Key: "reserved_instance_payment_option", DefaultValue: "", ValueType: schema.String},
	}
}

func (r *ElastiCacheCluster) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *ElastiCacheCluster) BuildResource() *schema.Resource {
	if r.HasReplicationGroup {
		return &schema.Resource{
			Name:        r.Address,
			NoPrice:     true,
			IsSkipped:   true,
			UsageSchema: r.UsageSchema(),
		}
	}

	costComponents := []*schema.CostComponent{
		r.elastiCacheCostComponent(false),
	}

	if strings.ToLower(r.Engine) == "redis" && r.SnapshotRetentionLimit > 1 {
		costComponents = append(costComponents, r.backupStorageCostComponent())
	}

	return &schema.Resource{
		Name:           r.Address,
		CostComponents: costComponents,
	}
}

func (r *ElastiCacheCluster) elastiCacheCostComponent(autoscaling bool) *schema.CostComponent {
	purchaseOptionLabel := "on-demand"
	priceFilter := &schema.PriceFilter{
		PurchaseOption: strPtr("on_demand"),
	}
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

	nameParams := []string{purchaseOptionLabel, r.NodeType}
	if autoscaling {
		nameParams = append(nameParams, "autoscaling")
	}

	return &schema.CostComponent{
		Name:           fmt.Sprintf("ElastiCache (%s)", strings.Join(nameParams, ", ")),
		Unit:           "hours",
		UnitMultiplier: decimal.NewFromInt(1),
		HourlyQuantity: decimalPtr(decimal.NewFromInt(r.CacheNodes)),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(r.Region),
			Service:       strPtr("AmazonElastiCache"),
			ProductFamily: strPtr("Cache Instance"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "instanceType", Value: strPtr(r.NodeType)},
				{Key: "locationType", Value: strPtr("AWS Region")},
				{Key: "cacheEngine", Value: strPtr(cases.Title(language.English).String(r.Engine))},
			},
		},
		PriceFilter: priceFilter,
		UsageBased:  autoscaling,
	}
}

func (r *ElastiCacheCluster) backupStorageCostComponent() *schema.CostComponent {
	var monthlyBackupStorageGB *decimal.Decimal

	backupRetention := r.SnapshotRetentionLimit - 1

	if r.SnapshotStorageSizeGB != nil {
		snapshotSize := decimal.NewFromFloat(*r.SnapshotStorageSizeGB)
		monthlyBackupStorageGB = decimalPtr(snapshotSize.Mul(decimal.NewFromInt(backupRetention)))
	}

	return &schema.CostComponent{
		Name:            "Backup storage",
		Unit:            "GB",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: monthlyBackupStorageGB,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(r.Region),
			Service:       strPtr("AmazonElastiCache"),
			ProductFamily: strPtr("Storage Snapshot"),
		},
		UsageBased: true,
	}
}

type elasticacheReservationResolver struct {
	term          string
	paymentOption string
	cacheNodeType string
}

func (r elasticacheReservationResolver) isElasticacheReservedNodeLegacyOffering() bool {
	for k := range elasticacheReservedNodeCacheLegacyOfferings {
		if k == r.paymentOption {
			return true
		}
	}
	return false
}

// PriceFilter implementation for elasticacheReservationResolver
// Allowed values for ReservedInstanceTerm: ["1_year", "3_year"]
// Allowed values for ReservedInstancePaymentOption: ["all_upfront", "partial_upfront", "no_upfront"] for non legacy reservation nodes
// Allowed values for ReservedInstancePaymentOption: ["heavy_utilization", "medium_utilization", "light_utilization"] for legacy reservation nodes
// Legacy reservation nodes: "t2", "m3", "m4", "r3", "r4". (See elasticacheReservedNodeLegacyTypes in util.go)
// Corner Case: In the case of legacy reservation cache nodes unfortunately, for a specified node type, the allowed ReservedInstancePaymentOption may differ in different regions.
//
//	Because of this, in the case of a legacy reservation, a warning is raised to the user.
func (r elasticacheReservationResolver) PriceFilter() (*schema.PriceFilter, error) {
	purchaseOptionLabel := "reserved"
	def := &schema.PriceFilter{
		PurchaseOption: strPtr(purchaseOptionLabel),
	}
	termLength := reservedTermsMapping[r.term]
	var purchaseOption string
	if r.isElasticacheReservedNodeLegacyOffering() {
		purchaseOption = elasticacheReservedNodeCacheLegacyOfferings[r.paymentOption]
	} else {
		purchaseOption = reservedPaymentOptionMapping[r.paymentOption]
	}
	validTerms := sliceOfKeysFromMap(reservedTermsMapping)
	if !stringInSlice(validTerms, r.term) {
		return def, fmt.Errorf("Invalid reserved_instance_term, ignoring reserved options. Expected: %s. Got: %s", strings.Join(validTerms, ", "), r.term)
	}
	validOptions := append(sliceOfKeysFromMap(reservedPaymentOptionMapping), sliceOfKeysFromMap(elasticacheReservedNodeCacheLegacyOfferings)...)

	if !stringInSlice(validOptions, r.paymentOption) {
		return def, fmt.Errorf("Invalid reserved_instance_payment_option, ignoring reserved options. Expected: %s. Got: %s", strings.Join(validOptions, ", "), r.paymentOption)
	}
	nodeType := strings.Split(r.cacheNodeType, ".")[1] // Get node type from cache node type. cache.m3.large -> m3
	if stringInSlice(elasticacheReservedNodeLegacyTypes, nodeType) {
		logging.Logger.Warn().Msgf("No products found is possible for legacy nodes %s if provided payment option is not supported by the region.", strings.Join(elasticacheReservedNodeLegacyTypes, ", "))
	}
	return &schema.PriceFilter{
		PurchaseOption:     strPtr(purchaseOptionLabel),
		StartUsageAmount:   strPtr("0"),
		TermLength:         strPtr(termLength),
		TermPurchaseOption: strPtr(purchaseOption),
	}, nil
}
