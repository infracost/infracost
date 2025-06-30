package azure_test

import (
	"testing"

	"github.com/infracost/infracost/internal/providers/pulumi/putest"
	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/testutil"
)

func TestStorageAccount(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	// Example Pulumi preview JSON with a storage account
	pulumiJSON := `{
		"steps": [{
			"resource": {
				"type": "azure:storage/account:Account",
				"name": "mystorageaccount",
				"urn": "urn:pulumi:dev::test::azure:storage/account:Account::mystorageaccount",
				"properties": {
					"name": "mystorageaccount",
					"resourceGroupName": "storage-rg",
					"location": "eastus",
					"accountTier": "Standard",
					"accountReplicationType": "LRS",
					"minTlsVersion": "TLS1_2",
					"enableHttpsTrafficOnly": true,
					"allowBlobPublicAccess": false
				}
			}
		}]
	}`

	usage := schema.NewUsageMap(map[string]interface{}{
		"mystorageaccount": map[string]interface{}{
			"storage_gb": map[string]interface{}{
				"hot":   1000,
				"cool":  2000,
				"archive": 5000,
			},
			"monthly_hot_block_blob_data_retrieval_gb": 100,
			"monthly_cool_block_blob_data_retrieval_gb": 200,
			"monthly_archive_block_blob_data_retrieval_gb": 500,
			"monthly_class_a_operations": 10000000,
			"monthly_class_b_operations": 20000000,
			"monthly_data_write_operations": 5000000,
		},
	})

	resourceChecks := []testutil.ResourceCheck{
		{
			Name: "azurerm_storage_account.mystorageaccount",
			CostComponentChecks: []testutil.CostComponentCheck{
				{Name: "Hot block blob storage", MonthlyQuantity: testutil.FloatPtr(1000)},
				{Name: "Cool block blob storage", MonthlyQuantity: testutil.FloatPtr(2000)},
				{Name: "Archive block blob storage", MonthlyQuantity: testutil.FloatPtr(5000)},
				{Name: "Hot block blob data retrieval", MonthlyQuantity: testutil.FloatPtr(100)},
				{Name: "Cool block blob data retrieval", MonthlyQuantity: testutil.FloatPtr(200)},
				{Name: "Archive block blob data retrieval", MonthlyQuantity: testutil.FloatPtr(500)},
				{Name: "Class A operations", MonthlyQuantity: testutil.FloatPtr(10)},
				{Name: "Class B operations", MonthlyQuantity: testutil.FloatPtr(20)},
				{Name: "Data write operations", MonthlyQuantity: testutil.FloatPtr(5)},
			},
		},
	}

	putest.ResourceTests(t, pulumiJSON, usage, resourceChecks)
}