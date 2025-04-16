package azure

import (
	"github.com/tidwall/gjson"

	"github.com/infracost/infracost/internal/schema"
)

func GetAzureRMHDInsightInteractiveQueryClusterRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "azurerm_hdinsight_interactive_query_cluster", //nolint:misspell
		RFunc: NewAzureRMHDInsightInteractiveQueryCluster,
	}
}

func NewAzureRMHDInsightInteractiveQueryCluster(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	region := d.Region

	costComponents := []*schema.CostComponent{}

	headNodeVM := d.Get("roles.0.headNode.0.vmSize").String()
	workerNodeVM := d.Get("roles.0.workerNode.0.vmSize").String()
	var workerInstances int64
	if d.Get("roles.0.workerNode.0.targetInstanceCount").Type != gjson.Null {
		workerInstances = d.Get("roles.0.workerNode.0.targetInstanceCount").Int()
	}
	zookeeperNodeVM := d.Get("roles.0.zookeeperNode.0.vmSize").String()

	costComponents = append(costComponents, hdInsightVMCostComponent(region, "Head", headNodeVM, 2))
	costComponents = append(costComponents, hdInsightVMCostComponent(region, "Worker", workerNodeVM, workerInstances))
	costComponents = append(costComponents, hdInsightVMCostComponent(region, "Zookeeper", zookeeperNodeVM, 3))

	return &schema.Resource{
		Name:           d.Address,
		CostComponents: costComponents,
	}
}
