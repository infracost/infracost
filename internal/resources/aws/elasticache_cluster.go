package aws

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"
	log "github.com/sirupsen/logrus"

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

var ElastiCacheClusterUsageSchema = []*schema.UsageItem{
	{Key: "snapshot_storage_size_gb", ValueType: schema.Float64, DefaultValue: 0},
	{Key: "reserved_instance_term", DefaultValue: "", ValueType: schema.String},
	{Key: "reserved_instance_payment_option", DefaultValue: "", ValueType: schema.String},
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
			UsageSchema: ElastiCacheClusterUsageSchema,
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
		resolver := reservedInstanceResolver{
			term:          strVal(r.ReservedInstanceTerm),
			paymentOption: strVal(r.ReservedInstancePaymentOption),
		}
		valid, err := resolver.Validate()
		if err != "" {
			log.Warnf(err)
		}
		if valid {
			purchaseOptionLabel = "reserved"
			priceFilter = &schema.PriceFilter{
				PurchaseOption:     strPtr("reserved"),
				StartUsageAmount:   strPtr("0"),
				TermLength:         strPtr(resolver.Term()),
				TermPurchaseOption: strPtr(resolver.PaymentOption()),
			}
		}
	}

	nameParams := []string{purchaseOptionLabel, r.NodeType}
	if autoscaling {
		nameParams = append(nameParams, "autoscaling")
	}
	ignoreIfMissingPrice := isElasticacheReservedNodeLegacyOffering(strVal(r.ReservedInstancePaymentOption))

	return &schema.CostComponent{
		Name:                 fmt.Sprintf("ElastiCache (%s)", strings.Join(nameParams, ", ")),
		Unit:                 "hours",
		UnitMultiplier:       decimal.NewFromInt(1),
		HourlyQuantity:       decimalPtr(decimal.NewFromInt(r.CacheNodes)),
		IgnoreIfMissingPrice: ignoreIfMissingPrice,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(r.Region),
			Service:       strPtr("AmazonElastiCache"),
			ProductFamily: strPtr("Cache Instance"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "instanceType", Value: strPtr(r.NodeType)},
				{Key: "locationType", Value: strPtr("AWS Region")},
				{Key: "cacheEngine", Value: strPtr(strings.Title(r.Engine))},
			},
		},
		PriceFilter: priceFilter,
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
	}
}
