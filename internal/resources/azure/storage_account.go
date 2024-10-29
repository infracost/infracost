package azure

import (
	"fmt"
	"strings"

	"github.com/shopspring/decimal"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"github.com/infracost/infracost/internal/logging"
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/usage"
)

// StorageAccount represents Azure data storage services.
//
// More resource information here:
//
//	Block Blob Storage: https://azure.microsoft.com/en-us/services/storage/blobs/
//	File Storage: https://azure.microsoft.com/en-us/services/storage/files/
//
// Pricing information here:
//
//	Block Blob Storage: https://azure.microsoft.com/en-us/pricing/details/storage/blobs/
//	File Storage: https://azure.microsoft.com/en-us/pricing/details/storage/files/
type StorageAccount struct {
	Address string
	Region  string

	AccessTier             string
	AccountKind            string
	AccountReplicationType string
	AccountTier            string
	NFSv3                  bool

	// "usage" args
	MonthlyStorageGB                        *float64 `infracost_usage:"storage_gb"`
	MonthlyIterativeReadOperations          *int64   `infracost_usage:"monthly_iterative_read_operations"`
	MonthlyReadOperations                   *int64   `infracost_usage:"monthly_read_operations"`
	MonthlyIterativeWriteOperations         *int64   `infracost_usage:"monthly_iterative_write_operations"`
	MonthlyWriteOperations                  *int64   `infracost_usage:"monthly_write_operations"`
	MonthlyListAndCreateContainerOperations *int64   `infracost_usage:"monthly_list_and_create_container_operations"`
	MonthlyOtherOperations                  *int64   `infracost_usage:"monthly_other_operations"`
	MonthlyDataRetrievalGB                  *float64 `infracost_usage:"monthly_data_retrieval_gb"`
	MonthlyDataWriteGB                      *float64 `infracost_usage:"monthly_data_write_gb"`
	BlobIndexTags                           *int64   `infracost_usage:"blob_index_tags"`
	DataAtRestStorageGB                     *float64 `infracost_usage:"data_at_rest_storage_gb"`
	SnapshotsStorageGB                      *float64 `infracost_usage:"snapshots_storage_gb"`
	MetadataAtRestStorageGB                 *float64 `infracost_usage:"metadata_at_rest_storage_gb"`
	EarlyDeletionGB                         *float64 `infracost_usage:"early_deletion_gb"`
}

// CoreType returns the name of this resource type
func (r *StorageAccount) CoreType() string {
	return "StorageAccount"
}

func (r *StorageAccount) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{
		{Key: "storage_gb", DefaultValue: 0, ValueType: schema.Float64},
		{Key: "monthly_iterative_read_operations", DefaultValue: 0, ValueType: schema.Int64},
		{Key: "monthly_read_operations", DefaultValue: 0, ValueType: schema.Int64},
		{Key: "monthly_iterative_write_operations", DefaultValue: 0, ValueType: schema.Int64},
		{Key: "monthly_write_operations", DefaultValue: 0, ValueType: schema.Int64},
		{Key: "monthly_list_and_create_container_operations", DefaultValue: 0, ValueType: schema.Int64},
		{Key: "monthly_other_operations", DefaultValue: 0, ValueType: schema.Int64},
		{Key: "monthly_data_retrieval_gb", DefaultValue: 0, ValueType: schema.Float64},
		{Key: "monthly_data_write_gb", DefaultValue: 0, ValueType: schema.Float64},
		{Key: "blob_index_tags", DefaultValue: 0, ValueType: schema.Int64},
		{Key: "data_at_rest_storage_gb", DefaultValue: 0, ValueType: schema.Float64},
		{Key: "snapshots_storage_gb", DefaultValue: 0, ValueType: schema.Float64},
		{Key: "metadata_at_rest_storage_gb", DefaultValue: 0, ValueType: schema.Float64},
		{Key: "early_deletion_gb", DefaultValue: 0, ValueType: schema.Float64},
	}
}

// PopulateUsage parses the u schema.UsageData into the StorageAccount.
// It uses the `infracost_usage` struct tags to populate data into the StorageAccount.
func (r *StorageAccount) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

// BuildResource builds a schema.Resource from valid StorageAccount data.
// This method is called after the resource is initialized by an IaC provider.
func (r *StorageAccount) BuildResource() *schema.Resource {
	if !r.isReplicationTypeSupported() {
		logging.Logger.Warn().Msgf("Skipping resource %s. %s %s doesn't support %s redundancy", r.Address, r.AccountKind, r.AccountTier, r.AccountReplicationType)
		return nil
	}

	if r.isPremium() {
		// Premium tier doesn't differentiate between Hot or Cool storage. This
		// helps to simplify skuName search.
		r.AccessTier = "Premium"
	}

	if r.isStorageV1() {
		// StorageV1 doesn't support Hot or Cool storage.
		r.AccessTier = "Standard"
	}

	costComponents := []*schema.CostComponent{}

	costComponents = append(costComponents, r.storageCostComponents()...)

	costComponents = append(costComponents, r.dataAtRestCostComponents()...)
	costComponents = append(costComponents, r.snapshotsCostComponents()...)
	costComponents = append(costComponents, r.metadataAtRestCostComponents()...)

	costComponents = append(costComponents, r.iterativeWriteOperationsCostComponents()...)
	costComponents = append(costComponents, r.writeOperationsCostComponents()...)
	costComponents = append(costComponents, r.listAndCreateContainerOperationsCostComponents()...)
	costComponents = append(costComponents, r.iterativeReadOperationsCostComponents()...)
	costComponents = append(costComponents, r.readOperationsCostComponents()...)
	costComponents = append(costComponents, r.otherOperationsCostComponents()...)
	costComponents = append(costComponents, r.dataRetrievalCostComponents()...)
	costComponents = append(costComponents, r.dataWriteCostComponents()...)
	costComponents = append(costComponents, r.blobIndexTagsCostComponents()...)

	costComponents = append(costComponents, r.earlyDeletionCostComponents()...)

	return &schema.Resource{
		Name:           r.Address,
		UsageSchema:    r.UsageSchema(),
		CostComponents: costComponents,
	}
}

// buildProductFilter returns a product filter for the Storage Account's products.
func (r *StorageAccount) buildProductFilter(meterName string) *schema.ProductFilter {
	var productName string

	switch {
	case r.isBlockBlobStorage():
		productName = map[string]string{
			"Standard": "Blob Storage",
			"Premium":  "Premium Block Blob",
		}[r.AccountTier]
	case r.isStorageV1():
		productName = map[string]string{
			"Standard": "General Block Blob",
			"Premium":  "Premium Block Blob",
		}[r.AccountTier]
	case r.isStorageV2():
		if r.NFSv3 {
			productName = map[string]string{
				"Standard": "General Block Blob v2 Hierarchical Namespace",
				"Premium":  "Premium Block Blob v2 Hierarchical Namespace",
			}[r.AccountTier]
		} else if strings.EqualFold(r.AccountReplicationType, "lrs") && r.isHot() {
			// For some reason the Azure pricing doesn't contain all the LRS costs for all regions under "General Block Blob v2" product name.
			// But, the same pricing is available under "Blob Storage" product name.
			productName = map[string]string{
				"Standard": "Blob Storage",
				"Premium":  "Premium Block Blob",
			}[r.AccountTier]
		} else {
			productName = map[string]string{
				"Standard": "General Block Blob v2",
				"Premium":  "Premium Block Blob",
			}[r.AccountTier]
		}
	case r.isBlobStorage():
		productName = map[string]string{
			"Standard": "Blob Storage",
			"Premium":  "Premium Block Blob",
		}[r.AccountTier]
	case r.isFileStorage():
		productName = map[string]string{
			"Standard": "Files v2",
			"Premium":  "Premium Files",
		}[r.AccountTier]
	}

	skuName := fmt.Sprintf("%s %s", cases.Title(language.English).String(r.AccessTier), strings.ToUpper(r.AccountReplicationType))

	return &schema.ProductFilter{
		VendorName:    strPtr("azure"),
		Region:        strPtr(r.Region),
		Service:       strPtr("Storage"),
		ProductFamily: strPtr("Storage"),
		AttributeFilters: []*schema.AttributeFilter{
			{Key: "productName", Value: strPtr(productName)},
			{Key: "skuName", Value: strPtr(skuName)},
			{Key: "meterName", ValueRegex: regexPtr(fmt.Sprintf("%s$", meterName))},
		},
	}
}

// storageCostComponents returns one or several tier cost components for monthly
// storage capacity in Blob Storage.
//
// BlockBlobStorage:
//
//	Standard Hot:  cost exists
//	Standard Cool: cost exists
//	Premium:       cost exists
//
// BlobStorage:
//
//	Standard Hot:  cost exists
//	Standard Cool: cost exists
//	Premium:       cost exists
//
// Storage:
//
// Standard: cost exists
//
// StorageV2:
//
//	Standard Hot:        cost exists
//	Standard Hot NFSv3:  cost exists
//	Standard Cool:       cost exists
//	Standard Cool NFSv3: cost exists
//	Premium:             cost exists
//	Premium NFSv3:       cost exists
//
// FileStorage: see dataAtRestCostComponents()
func (r *StorageAccount) storageCostComponents() []*schema.CostComponent {
	costComponents := []*schema.CostComponent{}

	if r.isFileStorage() {
		return costComponents
	}

	var quantity *decimal.Decimal
	name := "Capacity"

	if r.MonthlyStorageGB == nil {
		costComponents = append(costComponents, r.buildStorageCostComponent(
			name,
			"0",
			quantity,
		))
		return costComponents
	}

	quantity = decimalPtr(decimal.NewFromFloat(*r.MonthlyStorageGB))

	// Only Hot storage has pricing tiers, others have a single price for any
	// amount.
	if !r.isHot() {
		costComponents = append(costComponents, r.buildStorageCostComponent(
			name,
			"0",
			quantity,
		))
		return costComponents
	}

	type dataTier struct {
		name       string
		startUsage string
	}

	data := []dataTier{
		{name: fmt.Sprintf("%s (first 50TB)", name), startUsage: "0"},
		{name: fmt.Sprintf("%s (next 450TB)", name), startUsage: "51200"},
		{name: fmt.Sprintf("%s (over 500TB)", name), startUsage: "512000"},
	}

	tierLimits := []int{51200, 512000}
	tiers := usage.CalculateTierBuckets(*quantity, tierLimits)

	for i, d := range data {
		if i < len(tiers) && tiers[i].GreaterThan(decimal.Zero) {
			costComponents = append(costComponents, r.buildStorageCostComponent(
				d.name,
				d.startUsage,
				decimalPtr(tiers[i]),
			))
		}
	}

	return costComponents
}

// iterativeWriteOperationsCostComponents returns a cost component for Iterative
// Write Operations.
//
// BlockBlobStorage: n/a
//
// BlobStorage: n/a
//
// Storage: n/a
//
// StorageV2:
//
//	Standard Hot:        no cost
//	Standard Hot NFSv3:  cost exists
//	Standard Cool:       no cost
//	Standard Cool NFSv3: cost exists
//	Premium:             no cost
//	Premium NFSv3:       no cost
//
// FileStorage: n/a
func (r *StorageAccount) iterativeWriteOperationsCostComponents() []*schema.CostComponent {
	costComponents := []*schema.CostComponent{}

	if !r.isStorageV2() || !r.NFSv3 || r.isPremium() {
		return costComponents
	}

	var quantity *decimal.Decimal
	itemsPerCost := 100

	if r.MonthlyIterativeWriteOperations != nil {
		value := decimal.NewFromInt(*r.MonthlyIterativeWriteOperations)
		quantity = decimalPtr(value.Div(decimal.NewFromInt(int64(itemsPerCost))))
	}

	meterName := "Iterative Write Operations"

	costComponents = append(costComponents, &schema.CostComponent{
		Name:                 "Iterative write operations",
		Unit:                 "100 operations",
		UnitMultiplier:       decimal.NewFromInt(1),
		MonthlyQuantity:      quantity,
		IgnoreIfMissingPrice: r.canSkipPrice(),
		ProductFilter:        r.buildProductFilter(meterName),
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr("Consumption"),
		},
		UsageBased: true,
	})

	return costComponents
}

// writeOperationsCostComponents returns a cost component for Write Operations.
//
// BlockBlobStorage:
//
//	Standard Hot:  cost exists
//	Standard Cool: cost exists
//	Premium:       cost exists
//
// BlobStorage:
//
//	Standard Hot:  cost exists
//	Standard Cool: cost exists
//	Premium:       cost exists
//
// Storage:
//
// Standard: cost exists
//
// StorageV2:
//
//	Standard Hot:        cost exists
//	Standard Hot NFSv3:  cost exists
//	Standard Cool:       cost exists
//	Standard Cool NFSv3: cost exists
//	Premium:             cost exists
//	Premium NFSv3:       cost exists
//
// FileStorage:
//
//	Standard Hot:  cost exists
//	Standard Cool: cost exists
//	Premium:       no cost
func (r *StorageAccount) writeOperationsCostComponents() []*schema.CostComponent {
	costComponents := []*schema.CostComponent{}

	if r.isFileStorage() && r.isPremium() {
		return costComponents
	}

	var quantity *decimal.Decimal
	itemsPerCost := 10000

	if r.MonthlyWriteOperations != nil {
		value := decimal.NewFromInt(*r.MonthlyWriteOperations)
		quantity = decimalPtr(value.Div(decimal.NewFromInt(int64(itemsPerCost))))
	}

	meterName := "Write Operations"
	if r.isStorageV2() && r.NFSv3 {
		meterName = "(?<!Iterative) Write Operations"
	}

	costComponents = append(costComponents, &schema.CostComponent{
		Name:                 "Write operations",
		Unit:                 "10k operations",
		UnitMultiplier:       decimal.NewFromInt(1),
		MonthlyQuantity:      quantity,
		IgnoreIfMissingPrice: r.canSkipPrice(),
		ProductFilter:        r.buildProductFilter(meterName),
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr("Consumption"),
		},
		UsageBased: true,
	})

	return costComponents
}

// listAndCreateContainerOperationsCostComponents returns a cost component for
// List and Create Container Operations (List Operations for File storage).
//
// BlockBlobStorage:
//
//	Standard Hot:  no cost
//	Standard Cool: cost exists
//	Premium:       cost exists
//
// BlobStorage:
//
//	Standard Hot:  cost exists
//	Standard Cool: cost exists
//	Premium:       cost exists
//
// Storage:
//
// Standard: cost exists
//
// StorageV2:
//
//	Standard Hot:        cost exists
//	Standard Hot NFSv3:  no cost
//	Standard Cool:       cost exists
//	Standard Cool NFSv3: no cost
//	Premium:             no cost
//	Premium NFSv3:       no cost
//
// FileStorage:
//
//	Standard Hot:  cost exists
//	Standard Cool: cost exists
//	Premium:       no cost
func (r *StorageAccount) listAndCreateContainerOperationsCostComponents() []*schema.CostComponent {
	costComponents := []*schema.CostComponent{}

	if r.isFileStorage() && r.isPremium() {
		return costComponents
	}

	if r.isStorageV2() && r.NFSv3 {
		return costComponents
	}

	var quantity *decimal.Decimal
	itemsPerCost := 10000

	if r.MonthlyListAndCreateContainerOperations != nil {
		value := decimal.NewFromInt(*r.MonthlyListAndCreateContainerOperations)
		quantity = decimalPtr(value.Div(decimal.NewFromInt(int64(itemsPerCost))))
	}

	name := "List and create container operations"
	meterName := "List and Create Container Operations"

	if r.isFileStorage() {
		name = "List operations"
		meterName = "List Operations"
	}

	costComponents = append(costComponents, &schema.CostComponent{
		Name:                 name,
		Unit:                 "10k operations",
		UnitMultiplier:       decimal.NewFromInt(1),
		MonthlyQuantity:      quantity,
		IgnoreIfMissingPrice: r.canSkipPrice(),
		ProductFilter:        r.buildProductFilter(meterName),
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr("Consumption"),
		},
		UsageBased: true,
	})

	return costComponents
}

// iterativeReadOperationsCostComponents returns a cost component for Iterative Read Operations.
//
// BlockBlobStorage: n/a
//
// BlobStorage: n/a
//
// Storage: n/a
//
// StorageV2:
//
//	Standard Hot:        no cost
//	Standard Hot NFSv3:  cost exists
//	Standard Cool:       no cost
//	Standard Cool NFSv3: cost exists
//	Premium:             no cost
//	Premium NFSv3:       no cost
//
// FileStorage: n/a
func (r *StorageAccount) iterativeReadOperationsCostComponents() []*schema.CostComponent {
	costComponents := []*schema.CostComponent{}

	if !r.isStorageV2() || !r.NFSv3 || r.isPremium() {
		return costComponents
	}

	var quantity *decimal.Decimal
	itemsPerCost := 10000

	if r.MonthlyIterativeReadOperations != nil {
		value := decimal.NewFromInt(*r.MonthlyIterativeReadOperations)
		quantity = decimalPtr(value.Div(decimal.NewFromInt(int64(itemsPerCost))))
	}

	meterName := "Iterative Read Operations"

	costComponents = append(costComponents, &schema.CostComponent{
		Name:                 "Iterative read operations",
		Unit:                 "10k operations",
		UnitMultiplier:       decimal.NewFromInt(1),
		MonthlyQuantity:      quantity,
		IgnoreIfMissingPrice: r.canSkipPrice(),
		ProductFilter:        r.buildProductFilter(meterName),
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr("Consumption"),
		},
		UsageBased: true,
	})

	return costComponents
}

// readOperationsCostComponents returns a cost component for Read Operations.
//
// BlockBlobStorage:
//
//	Standard Hot:  cost exists
//	Standard Cool: cost exists
//	Premium:       cost exists
//
// Storage:
//
// Standard: cost exists
//
// StorageV2:
//
//	Standard Hot:        cost exists
//	Standard Hot NFSv3:  cost exists
//	Standard Cool:       cost exists
//	Standard Cool NFSv3: cost exists
//	Premium:             cost exists
//	Premium NFSv3:       cost exists
//
// FileStorage:
//
//	Standard Hot:  cost exists
//	Standard Cool: cost exists
//	Premium:       no cost
func (r *StorageAccount) readOperationsCostComponents() []*schema.CostComponent {
	costComponents := []*schema.CostComponent{}

	if r.isFileStorage() && r.isPremium() {
		return costComponents
	}

	var quantity *decimal.Decimal
	itemsPerCost := 10000

	if r.MonthlyReadOperations != nil {
		value := decimal.NewFromInt(*r.MonthlyReadOperations)
		quantity = decimalPtr(value.Div(decimal.NewFromInt(int64(itemsPerCost))))
	}

	meterName := "Read Operations"
	if r.isStorageV2() && r.NFSv3 {
		meterName = "(?<!Iterative) Read Operations"
	}
	if r.isStorageV1() && contains([]string{"LRS", "GRS", "RA-GRS"}, strings.ToUpper(r.AccountReplicationType)) {
		// Storage V1 GRS/LRS/RA-GRS doesn't always have a Read Operations meter name, but we can use this regex
		// to match Read or Other Operations meter since they are the same price.
		meterName = "(Other|Read) Operations"
	}

	filter := r.buildProductFilter(meterName)
	costComponents = append(costComponents, &schema.CostComponent{
		Name:                 "Read operations",
		Unit:                 "10k operations",
		UnitMultiplier:       decimal.NewFromInt(1),
		MonthlyQuantity:      quantity,
		IgnoreIfMissingPrice: r.canSkipPrice(),
		ProductFilter:        filter,
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr("Consumption"),
		},
		UsageBased: true,
	})

	return costComponents
}

// otherOperationsCostComponents returns a cost component for All Other Operations.
//
// BlockBlobStorage:
//
//	Standard Hot:  cost exists
//	Standard Cool: cost exists
//	Premium:       cost exists
//
// BlobStorage:
//
//	Standard Hot:  cost exists
//	Standard Cool: cost exists
//	Premium:       cost exists
//
// Storage:
//
// Standard: cost exists
//
// StorageV2:
//
//	Standard Hot:        cost exists
//	Standard Hot NFSv3:  cost exists
//	Standard Cool:       cost exists
//	Standard Cool NFSv3: cost exists
//	Premium:             cost exists
//	Premium NFSv3:       cost exists
//
// FileStorage:
//
//	Standard Hot:  cost exists
//	Standard Cool: cost exists
//	Premium:       no cost
func (r *StorageAccount) otherOperationsCostComponents() []*schema.CostComponent {
	costComponents := []*schema.CostComponent{}

	if r.isFileStorage() && r.isPremium() {
		return costComponents
	}

	var quantity *decimal.Decimal
	itemsPerCost := 10000

	if r.MonthlyOtherOperations != nil {
		value := decimal.NewFromInt(*r.MonthlyOtherOperations)
		quantity = decimalPtr(value.Div(decimal.NewFromInt(int64(itemsPerCost))))
	}

	meterName := "Other Operations"
	if r.isStorageV1() {
		// Most StorageV1 rows don't have a meter name called Other Operations,
		// but they do have Delete Operations which is the same price.
		meterName = "Delete Operations"
	}

	costComponents = append(costComponents, &schema.CostComponent{
		Name:                 "All other operations",
		Unit:                 "10k operations",
		UnitMultiplier:       decimal.NewFromInt(1),
		MonthlyQuantity:      quantity,
		IgnoreIfMissingPrice: r.canSkipPrice(),
		ProductFilter:        r.buildProductFilter(meterName),
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr("Consumption"),
		},
		UsageBased: true,
	})

	return costComponents
}

// dataRetrievalCostComponents returns a cost component for Data Retrieval
// amount.
//
// BlockBlobStorage:
//
//	Standard Hot:  no cost
//	Standard Cool: cost exists
//	Premium:       no cost
//
// BlobStorage:
//
//	Standard Hot:  no cost
//	Standard Cool: cost exists
//	Premium:       no cost
//
// Storage: n/a
//
// StorageV2:
//
//	Standard Hot:        no cost
//	Standard Hot NFSv3:  no cost
//	Standard Cool:       cost exists
//	Standard Cool NFSv3: cost exists
//	Premium:             no cost
//	Premium NFSv3:       no cost
//
// FileStorage:
//
//	Standard Hot:  no cost
//	Standard Cool: cost exists
//	Premium:       no cost
func (r *StorageAccount) dataRetrievalCostComponents() []*schema.CostComponent {
	costComponents := []*schema.CostComponent{}

	if !r.isCool() {
		return costComponents
	}

	var quantity *decimal.Decimal

	if r.MonthlyDataRetrievalGB != nil {
		quantity = decimalPtr(decimal.NewFromFloat(*r.MonthlyDataRetrievalGB))
	}

	meterName := "Data Retrieval"

	costComponents = append(costComponents, &schema.CostComponent{
		Name:                 "Data retrieval",
		Unit:                 "GB",
		UnitMultiplier:       decimal.NewFromInt(1),
		MonthlyQuantity:      quantity,
		IgnoreIfMissingPrice: r.canSkipPrice(),
		ProductFilter:        r.buildProductFilter(meterName),
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr("Consumption"),
		},
		UsageBased: true,
	})

	return costComponents
}

// dataWriteCostComponents returns a cost component for Data Write amount.
//
// BlockBlobStorage:
//
//	Standard Hot:  no cost
//	Standard Cool: cost exists
//	Premium:       no cost
//
// BlobStorage:
//
//	Standard Hot:  no cost
//	Standard Cool: cost exists
//	Premium:       no cost
//
// Storage: n/a
//
// StorageV2:
//
//	Standard Hot:        no cost
//	Standard Hot NFSv3:  no cost
//	Standard Cool:       no cost
//	Standard Cool NFSv3: no cost
//	Premium:             no cost
//	Premium NFSv3:       no cost
//
// FileStorage:
//
//	Standard Hot:  no cost
//	Standard Cool: no cost
//	Premium:       no cost
func (r *StorageAccount) dataWriteCostComponents() []*schema.CostComponent {
	costComponents := []*schema.CostComponent{}

	if !(r.isBlockBlobStorage() && !r.isBlobStorage()) || !r.isCool() {
		return costComponents
	}

	var quantity *decimal.Decimal

	if r.MonthlyDataWriteGB != nil {
		quantity = decimalPtr(decimal.NewFromFloat(*r.MonthlyDataWriteGB))
	}

	meterName := "Data Write"

	costComponents = append(costComponents, &schema.CostComponent{
		Name:                 "Data write",
		Unit:                 "GB",
		UnitMultiplier:       decimal.NewFromInt(1),
		MonthlyQuantity:      quantity,
		IgnoreIfMissingPrice: r.canSkipPrice(),
		ProductFilter:        r.buildProductFilter(meterName),
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr("Consumption"),
		},
		UsageBased: true,
	})

	return costComponents
}

// blobIndexTagsCostComponents returns a cost component for Blob Index
// subresources amount.
//
// BlockBlobStorage:
//
//	Standard Hot:  cost exists
//	Standard Cool: cost exists
//	Premium:       no cost
//
// BlobStorage:
//
//	Standard Hot:  cost exists
//	Standard Cool: cost exists
//	Premium:       no cost
//
// Storage: n/a
//
// StorageV2:
//
//	Standard Hot:        cost exists
//	Standard Hot NFSv3:  no cost
//	Standard Cool:       cost exists
//	Standard Cool NFSv3: no cost
//	Premium:             no cost
//	Premium NFSv3:       no cost
//
// FileStorage:
//
//	Standard Hot:  no cost
//	Standard Cool: no cost
//	Premium:       no cost
func (r *StorageAccount) blobIndexTagsCostComponents() []*schema.CostComponent {
	costComponents := []*schema.CostComponent{}

	isBlockPremium := r.isBlockBlobStorage() && r.isPremium()
	isBlobPremium := r.isBlobStorage() && r.isPremium()
	isV2NFSv3 := r.isStorageV2() && (r.NFSv3 || r.isPremium())
	if r.isFileStorage() || r.isStorageV1() || isBlockPremium || isBlobPremium || isV2NFSv3 {
		return costComponents
	}

	var quantity *decimal.Decimal
	itemsPerCost := 10000

	if r.BlobIndexTags != nil {
		value := decimal.NewFromInt(*r.BlobIndexTags)
		quantity = decimalPtr(value.Div(decimal.NewFromInt(int64(itemsPerCost))))
	}

	meterName := "Index Tags"

	costComponents = append(costComponents, &schema.CostComponent{
		Name:                 "Blob index",
		Unit:                 "10k tags",
		UnitMultiplier:       decimal.NewFromInt(1),
		MonthlyQuantity:      quantity,
		IgnoreIfMissingPrice: r.canSkipPrice(),
		ProductFilter:        r.buildProductFilter(meterName),
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr("Consumption"),
		},
		UsageBased: true,
	})

	return costComponents
}

// dataAtRestCostComponents returns a cost component for Data at Rest amount in
// File Storage.
//
// BlockBlobStorage: n/a
//
// BlobStorage: n/a
//
// Storage: n/a
//
// StorageV2: n/a
//
// FileStorage:
//
//	Standard Hot:  cost exists
//	Standard Cool: cost exists
//	Premium:       cost exists
func (r *StorageAccount) dataAtRestCostComponents() []*schema.CostComponent {
	costComponents := []*schema.CostComponent{}

	if !r.isFileStorage() {
		return costComponents
	}

	var quantity *decimal.Decimal

	if r.DataAtRestStorageGB != nil {
		quantity = decimalPtr(decimal.NewFromFloat(*r.DataAtRestStorageGB))
	}

	meterName := "Data Stored"
	if r.isPremium() {
		meterName = "Provisioned"
	}

	costComponents = append(costComponents, &schema.CostComponent{
		Name:            "Data at rest",
		Unit:            "GB",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: quantity,
		ProductFilter:   r.buildProductFilter(meterName),
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr("Consumption"),
		},
		UsageBased: true,
	})

	return costComponents
}

// snapshotsCostComponents returns a cost component for Snapshots amount in
// File Storage.
//
// BlockBlobStorage: n/a
//
// BlobStorage: n/a
//
// Storage: n/a
//
// StorageV2: n/a
//
// FileStorage:
//
//	Standard Hot:  cost exists
//	Standard Cool: cost exists
//	Premium:       cost exists
func (r *StorageAccount) snapshotsCostComponents() []*schema.CostComponent {
	costComponents := []*schema.CostComponent{}

	if !r.isFileStorage() {
		return costComponents
	}

	var quantity *decimal.Decimal

	if r.SnapshotsStorageGB != nil {
		quantity = decimalPtr(decimal.NewFromFloat(*r.SnapshotsStorageGB))
	}

	meterName := "Data Stored"
	if r.isPremium() {
		meterName = "Snapshots"
	}

	costComponents = append(costComponents, &schema.CostComponent{
		Name:            "Snapshots",
		Unit:            "GB",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: quantity,
		ProductFilter:   r.buildProductFilter(meterName),
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr("Consumption"),
		},
		UsageBased: true,
	})

	return costComponents
}

// metadataAtRestCostComponents returns a cost component for Metadata at-rest amount in
// File Storage.
//
// BlockBlobStorage: n/a
//
// BlobStorage: n/a
//
// Storage: n/a
//
// StorageV2: n/a
//
// FileStorage:
//
//	Standard Hot:  cost exists
//	Standard Cool: cost exists
//	Premium:       no cost
func (r *StorageAccount) metadataAtRestCostComponents() []*schema.CostComponent {
	costComponents := []*schema.CostComponent{}

	if !r.isFileStorage() || r.isPremium() {
		return costComponents
	}

	var quantity *decimal.Decimal

	if r.MetadataAtRestStorageGB != nil {
		quantity = decimalPtr(decimal.NewFromFloat(*r.MetadataAtRestStorageGB))
	}

	meterName := "Metadata"

	costComponents = append(costComponents, &schema.CostComponent{
		Name:            "Metadata at rest",
		Unit:            "GB",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: quantity,
		ProductFilter:   r.buildProductFilter(meterName),
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr("Consumption"),
		},
		UsageBased: true,
	})

	return costComponents
}

// earlyDeletionCostComponents returns a cost component for Metadata at-rest amount in
// File Storage.
//
// BlockBlobStorage: n/a
//
// BlobStorage: n/a
//
// Storage: n/a
//
// StorageV2:
//
//	Standard Hot:        no cost
//	Standard Hot NFSv3:  no cost
//	Standard Cool:       cost exists
//	Standard Cool NFSv3: cost exists
//	Premium:             no cost
//	Premium NFSv3:       no cost
//
// FileStorage:
//
//	Standard Hot:  no cost
//	Standard Cool: cost exists
//	Premium:       no cost
func (r *StorageAccount) earlyDeletionCostComponents() []*schema.CostComponent {
	costComponents := []*schema.CostComponent{}

	if r.isStorageV1() || r.isBlockBlobStorage() || r.isBlobStorage() || !r.isCool() {
		return costComponents
	}

	var quantity *decimal.Decimal

	if r.EarlyDeletionGB != nil {
		quantity = decimalPtr(decimal.NewFromFloat(*r.EarlyDeletionGB))
	}

	meterName := "Early Delete"

	costComponents = append(costComponents, &schema.CostComponent{
		Name:                 "Early deletion",
		Unit:                 "GB",
		UnitMultiplier:       decimal.NewFromInt(1),
		MonthlyQuantity:      quantity,
		IgnoreIfMissingPrice: r.canSkipPrice(),
		ProductFilter:        r.buildProductFilter(meterName),
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr("Consumption"),
		},
		UsageBased: true,
	})

	return costComponents
}

// buildStorageCostComponent builds one cost component for storage amount costs.
func (r *StorageAccount) buildStorageCostComponent(name string, startUsage string, quantity *decimal.Decimal) *schema.CostComponent {
	meterName := "Data Stored"

	return &schema.CostComponent{
		Name:                 name,
		Unit:                 "GB",
		UnitMultiplier:       decimal.NewFromInt(1),
		MonthlyQuantity:      quantity,
		IgnoreIfMissingPrice: r.canSkipPrice(),
		ProductFilter:        r.buildProductFilter(meterName),
		PriceFilter: &schema.PriceFilter{
			PurchaseOption:   strPtr("Consumption"),
			StartUsageAmount: strPtr(startUsage),
		},
		UsageBased: true,
	}
}

func (r *StorageAccount) isBlockBlobStorage() bool {
	return strings.EqualFold(r.AccountKind, "blockblobstorage")
}

func (r *StorageAccount) isFileStorage() bool {
	return strings.EqualFold(r.AccountKind, "filestorage")
}

func (r *StorageAccount) isBlobStorage() bool {
	return strings.EqualFold(r.AccountKind, "blobstorage")
}

func (r *StorageAccount) isStorageV1() bool {
	return strings.EqualFold(r.AccountKind, "storage")
}

func (r *StorageAccount) isStorageV2() bool {
	return strings.EqualFold(r.AccountKind, "storagev2")
}

func (r *StorageAccount) isHot() bool {
	return strings.EqualFold(r.AccessTier, "hot")
}

func (r *StorageAccount) isCool() bool {
	return strings.EqualFold(r.AccessTier, "cool")
}

func (r *StorageAccount) isPremium() bool {
	return strings.EqualFold(r.AccountTier, "premium")
}

func (r *StorageAccount) isReplicationTypeSupported() bool {
	var validReplicationTypes []string

	switch {
	case r.isPremium():
		validReplicationTypes = []string{"LRS", "ZRS"}
	case r.isBlockBlobStorage():
		validReplicationTypes = []string{"LRS", "GRS", "RA-GRS"}
	case r.isStorageV1():
		validReplicationTypes = []string{"LRS", "ZRS", "GRS", "RA-GRS"}
	case r.isStorageV2():
		validReplicationTypes = []string{"LRS", "ZRS", "GRS", "RA-GRS", "GZRS", "RA-GZRS"}
	case r.isBlobStorage():
		validReplicationTypes = []string{"LRS", "GRS", "RA-GRS"}
	case r.isFileStorage():
		validReplicationTypes = []string{"LRS", "GRS", "ZRS"}
	}

	if validReplicationTypes != nil {
		return contains(validReplicationTypes, strings.ToUpper(r.AccountReplicationType))
	}

	return true
}

func (r *StorageAccount) canSkipPrice() bool {
	// Not all regions support GZRS/RA-GZRS redundancy types. Some operations miss
	// prices for specific regions.
	// Read more: https://docs.microsoft.com/en-us/azure/storage/common/storage-redundancy
	return r.isStorageV2()
}
