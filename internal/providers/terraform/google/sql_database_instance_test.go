package google_test

import (
	"testing"

	"github.com/infracost/infracost/internal/providers/terraform/tftest"
	"github.com/infracost/infracost/internal/schema"
	"github.com/infracost/infracost/internal/testutil"
	"github.com/shopspring/decimal"
)

func TestNewSQLInstance_SharedInstance(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tf := `
	  resource "google_sql_database" "database" {
		name     = "my-database"
		instance = google_sql_database_instance.my_sql_instance.name
	  }
	  
	  resource "google_sql_database_instance" "my_sql_instance" {
		name   = "my-database-instance"
		region = "us-central1"
		database_version = "POSTGRES_11"
		settings {
		  tier = "db-f1-micro"
		  availability_type = "ZONAL"
		}
	  
		deletion_protection  = "true"
	  }	  
	`

	usage := schema.NewEmptyUsageMap()

	resourceChecks := []testutil.ResourceCheck{
		{
			Name: "google_sql_database_instance.my_sql_instance",
			CostComponentChecks: []testutil.CostComponentCheck{
				{
					Name:            "Instance pricing",
					PriceHash:       "00652373a56ac9fb427a933a88235829-ef2cadbde566a742ff14834f883bcb8a",
					HourlyCostCheck: testutil.MonthlyPriceMultiplierCheck(decimal.NewFromInt(1)),
				},
			},
		},
	}
	tftest.ResourceTests(t, tf, usage, resourceChecks)
}
