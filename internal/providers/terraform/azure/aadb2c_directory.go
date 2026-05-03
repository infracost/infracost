package azure

import (
	"github.com/infracost/infracost/internal/resources/azure"
	"github.com/infracost/infracost/internal/schema"
)

func GetAzureRMAADB2CDirectoryRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "azurerm_aadb2c_directory",
		RFunc: NewAzureRMAADB2CDirectory,
	}
}

func NewAzureRMAADB2CDirectory(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	region := d.Region
	var monthlyActiveUsers *int64
	if u != nil && u.Get("monthly_active_users").Exists() {
		v := u.Get("monthly_active_users").Int()
		monthlyActiveUsers = &v
	}

	res := &azure.AADB2CDirectory{
		Address:            d.Address,
		Region:             region,
		MonthlyActiveUsers: monthlyActiveUsers,
	}
	res.PopulateUsage(u)
	return res.BuildResource()
} 