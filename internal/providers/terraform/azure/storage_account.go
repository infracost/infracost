package azure

import (
	"fmt"
	"strings"

	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/usage"
	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"
)

func GetAzureStorageAccountRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "azurerm_storage_account",
		RFunc: NewAzureStorageAccount,
	}
}

func NewAzureStorageAccount(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	var costComponents []*schema.CostComponent

	location := d.Get("location").String()
	accountKind := "StorageV2"
	if d.Get("account_kind").Type != gjson.Null {
		accountKind = d.Get("account_kind").String()
	}

	if accountKind != "BlockBlobStorage" {
		log.Warnf("This resource supports only BlockBlobStorage account kind")
		return nil
	}

	accountReplicationType := d.Get("account_replication_type").String()
	accountTier := d.Get("account_tier").String()
	accessTier := "Hot"
	if d.Get("access_tier").Type != gjson.Null {
		accessTier = d.Get("access_tier").String()
	}

	productName := map[string]string{
		"Standard": "Blob Storage",
		"Premium":  "Premium Block Blob",
	}[accountTier]

	validPremiumReplicationTypes := []string{"ZRS", "LRS"}
	validStandardReplicationTypes := []string{"LRS", "GRS", "RAGRS"}

	if accessTier == "Premium" {
		if !Contains(validPremiumReplicationTypes, accountReplicationType) || !Contains(validStandardReplicationTypes, accountReplicationType) {
			log.Warnf("%s redundancy does not supports for %s performance tier", accountReplicationType, accountTier)
		}
	}

	var capacity, writeOperations, listOperations, readOperations, otherOperations, dataRetrieval, dataWrite, blobIndex *decimal.Decimal

	if accountReplicationType == "RAGRS" {
		accountReplicationType = "RA-GRS"
	}

	skuName := fmt.Sprintf("%s %s", accessTier, accountReplicationType)
	if accountTier == "Premium" {
		skuName = fmt.Sprintf("%s %s", accountTier, accountReplicationType)
	}

	if accountReplicationType == "RA-GRS" {
		accountReplicationType = "GRS"
	}

	var dataMeterName string
	if accountTier == "Premium" {
		dataMeterName = fmt.Sprintf("%s Data Stored", accountReplicationType)
	} else {
		dataMeterName = fmt.Sprintf("%s Data Stored", skuName)
	}

	if u != nil && u.Get("storage_gb").Exists() {
		capacity = decimalPtr(decimal.NewFromInt(u.Get("storage_gb").Int()))

		if accessTier == "Hot" {
			dataStorageTiers := []int{51200, 512000}
			dataStorageQuantities := usage.CalculateTierBuckets(*capacity, dataStorageTiers)

			costComponents = append(costComponents, blobDataStorageCostComponent(location, "Capacity (first 50TB)", skuName, "0", productName, dataMeterName, &dataStorageQuantities[0]))
			if dataStorageQuantities[1].GreaterThan(decimal.Zero) {
				costComponents = append(costComponents, blobDataStorageCostComponent(location, "Capacity (next 450TB)", skuName, "51200", productName, dataMeterName, &dataStorageQuantities[1]))
			}
			if dataStorageQuantities[2].GreaterThan(decimal.Zero) {
				costComponents = append(costComponents, blobDataStorageCostComponent(location, "Capacity (over 500TB)", skuName, "512000", productName, dataMeterName, &dataStorageQuantities[2]))
			}
		} else {
			costComponents = append(costComponents, blobDataStorageCostComponent(location, "Capacity", skuName, "0", productName, dataMeterName, capacity))
		}
	} else {
		var unknown *decimal.Decimal

		costComponents = append(costComponents, blobDataStorageCostComponent(location, "Capacity", skuName, "0", productName, dataMeterName, unknown))
	}

	if u != nil && u.Get("monthly_write_operations").Exists() {
		writeOperations = decimalPtr(decimal.NewFromInt(u.Get("monthly_write_operations").Int()))
	}
	meterName := fmt.Sprintf("%s %s Write Operations", accessTier, accountReplicationType)
	if accountTier == "Premium" {
		meterName = meterName[strings.Index(meterName, " ")+1:]
	}
	costComponents = append(costComponents, blobOperationsCostComponent(location, "Write operations", "10K operations", skuName, meterName, productName, writeOperations, 10000))

	if u != nil && u.Get("monthly_list_and_create_container_operations").Exists() {
		listOperations = decimalPtr(decimal.NewFromInt(u.Get("monthly_list_and_create_container_operations").Int()))
	}
	meterName = fmt.Sprintf("%s List and Create Container Operations", accountReplicationType)
	costComponents = append(costComponents, blobOperationsCostComponent(location, "List and create container operations", "10K operations", skuName, meterName, productName, listOperations, 10000))

	if u != nil && u.Get("monthly_read_operations").Exists() {
		readOperations = decimalPtr(decimal.NewFromInt(u.Get("monthly_read_operations").Int()))
	}
	meterName = fmt.Sprintf("%s Read Operations", accessTier)
	if accountTier == "Premium" {
		meterName = fmt.Sprintf("%s Read Operations", accountReplicationType)
	}
	costComponents = append(costComponents, blobOperationsCostComponent(location, "Read operations", "10K operations", skuName, meterName, productName, readOperations, 10000))

	if u != nil && u.Get("monthly_other_operations").Exists() {
		otherOperations = decimalPtr(decimal.NewFromInt(u.Get("monthly_other_operations").Int()))
	}
	meterName = "All Other Operations"
	if accountTier == "Premium" {
		meterName = fmt.Sprintf("%s %s", accountReplicationType, meterName)
	}
	costComponents = append(costComponents, blobOperationsCostComponent(location, "All other operations", "10K operations", skuName, meterName, productName, otherOperations, 10000))

	if accountTier != "Premium" {
		if u != nil && u.Get("monthly_data_retrieval_gb").Exists() {
			dataRetrieval = decimalPtr(decimal.NewFromInt(u.Get("monthly_data_retrieval_gb").Int()))
		}
		meterName = fmt.Sprintf("%s Data Retrieval", accessTier)
		costComponents = append(costComponents, blobOperationsCostComponent(location, "Data retrieval", "GB", skuName, meterName, productName, dataRetrieval, 1))

		if u != nil && u.Get("monthly_data_write_gb").Exists() {
			dataWrite = decimalPtr(decimal.NewFromInt(u.Get("monthly_data_write_gb").Int()))
		}
		meterName = "Data Write"
		if accessTier == "Cool" {
			meterName = fmt.Sprintf("%s %s %s", accessTier, accountReplicationType, meterName)
		}
		costComponents = append(costComponents, blobOperationsCostComponent(location, "Data write", "GB", skuName, meterName, productName, dataWrite, 1))

		if u != nil && u.Get("blob_index_tags").Exists() {
			blobIndex = decimalPtr(decimal.NewFromInt(u.Get("blob_index_tags").Int()))
		}
		meterName = "Index Tags"
		if strings.Contains(skuName, "GRS") {
			meterName = "GRS " + meterName
		}
		costComponents = append(costComponents, blobOperationsCostComponent(location, "Blob index", "10K tags", skuName, meterName, productName, blobIndex, 10000))
	}

	return &schema.Resource{
		Name:           d.Address,
		CostComponents: costComponents,
	}
}

func blobDataStorageCostComponent(location, name, skuName, startUsage, productName, meterName string, quantity *decimal.Decimal) *schema.CostComponent {
	return &schema.CostComponent{
		Name:                 name,
		Unit:                 "GB-months",
		UnitMultiplier:       1,
		MonthlyQuantity:      quantity,
		IgnoreIfMissingPrice: true,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("azure"),
			Region:        strPtr(location),
			Service:       strPtr("Storage"),
			ProductFamily: strPtr("Storage"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "productName", Value: strPtr(productName)},
				{Key: "skuName", Value: strPtr(skuName)},
				{Key: "meterName", Value: strPtr(meterName)},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption:   strPtr("Consumption"),
			StartUsageAmount: strPtr(startUsage),
		},
	}
}

func blobOperationsCostComponent(location, name, unit, skuName, meterName, productName string, quantity *decimal.Decimal, multi int) *schema.CostComponent {
	if quantity != nil {
		quantity = decimalPtr(quantity.Div(decimal.NewFromInt(int64(multi))))
	}

	return &schema.CostComponent{
		Name:                 name,
		Unit:                 unit,
		UnitMultiplier:       1,
		MonthlyQuantity:      quantity,
		IgnoreIfMissingPrice: true,
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("azure"),
			Region:        strPtr(location),
			Service:       strPtr("Storage"),
			ProductFamily: strPtr("Storage"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "productName", Value: strPtr(productName)},
				{Key: "skuName", Value: strPtr(skuName)},
				{Key: "meterName", Value: strPtr(meterName)},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr("Consumption"),
		},
	}
}
