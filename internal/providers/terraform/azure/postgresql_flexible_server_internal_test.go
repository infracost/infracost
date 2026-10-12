package azure

import (
	"testing"

	"github.com/tidwall/gjson"

	"github.com/infracost/infracost/internal/schema"
)

func TestPostgreSQLFlexibleServerHighAvailabilityQuantities(t *testing.T) {
	d := schema.NewResourceData("azurerm_postgresql_flexible_server", "azurerm", "azurerm_postgresql_flexible_server.ha", nil, gjson.Parse(`{
		"sku_name": "GP_Standard_D2ads_v5",
		"storage_mb": 32768,
		"storage_type": "PremiumV2_LRS",
		"storage_iops": 5000,
		"storage_throughput": 200,
		"high_availability": [{"mode": "ZoneRedundant", "standby_availability_zone": "2"}]
	}`))

	got := map[string]string{}
	for _, c := range newPostgreSQLFlexibleServer(d).BuildResource().CostComponents {
		switch {
		case c.HourlyQuantity != nil:
			got[c.Name] = c.HourlyQuantity.String()
		case c.MonthlyQuantity != nil:
			got[c.Name] = c.MonthlyQuantity.String()
		}
	}

	want := map[string]string{
		"Compute (GP_Standard_D2ads_v5)": "2",
		"Storage":                        "64",
		"Provisioned IOPS":               "4000",
		"Provisioned throughput":         "150",
	}
	for name, q := range want {
		if got[name] != q {
			t.Errorf("%s: got quantity %q, want %q", name, got[name], q)
		}
	}
}
