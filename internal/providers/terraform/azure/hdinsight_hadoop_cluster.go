package azure

import (
	"fmt"
	"strings"

	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
	"github.com/tidwall/gjson"
)

func GetAzureRMHDInsightHadoopClusterRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "azurerm_hdinsight_hadoop_cluster", //nolint:misspell
		RFunc: NewAzureRMHDInsightHadoopCluster,
	}
}

func NewAzureRMHDInsightHadoopCluster(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	location := d.Get("location").String()
	costComponents := []*schema.CostComponent{}

	headNodeVM := d.Get("roles.0.head_node.0.vm_size").String()
	workerNodeVM := d.Get("roles.0.worker_node.0.vm_size").String()
	var workerInstances int64
	if d.Get("roles.0.worker_node.0.target_instance_count").Type != gjson.Null {
		workerInstances = d.Get("roles.0.worker_node.0.target_instance_count").Int()
	}
	zookeeperNodeVM := d.Get("roles.0.zookeeper_node.0.vm_size").String()

	costComponents = append(costComponents, hdInsightVMCostComponent(location, "Head", headNodeVM, 2))
	costComponents = append(costComponents, hdInsightVMCostComponent(location, "Worker", workerNodeVM, workerInstances))
	costComponents = append(costComponents, hdInsightVMCostComponent(location, "Zookeeper", zookeeperNodeVM, 3))

	var edgeNodeVM string
	if d.Get("roles.0.edge_node.0.vm_size").Type != gjson.Null {
		edgeNodeVM = d.Get("roles.0.edge_node.0.vm_size").String()
		var workerInstances int64
		if d.Get("roles.0.edge_node.0.target_instance_count").Type != gjson.Null {
			workerInstances = d.Get("roles.0.edge_node.0.target_instance_count").Int()
		}
		costComponents = append(costComponents, hdInsightVMCostComponent(location, "Edge", edgeNodeVM, workerInstances))
	}

	return &schema.Resource{
		Name:           d.Address,
		CostComponents: costComponents,
	}
}

func hdInsightVMCostComponent(location, node, instanceType string, instances int64) *schema.CostComponent {
	skuName := parseVMSKUName(instanceType)
	t := strings.Split(skuName, " ")
	if len(t) > 1 {
		dSeries := []string{"D1", "D2", "D3", "D4", "D5"}
		if Contains(dSeries, t[0]) {
			skuName = t[0]
		} else {
			version := strings.ToLower(t[1])
			skuName = t[0] + " " + version
		}
	}

	return &schema.CostComponent{
		Name:           fmt.Sprintf("%s node (%s)", node, skuName),
		Unit:           "hours",
		UnitMultiplier: 1,
		HourlyQuantity: decimalPtr(decimal.NewFromInt(instances)),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("azure"),
			Region:        strPtr(location),
			Service:       strPtr("HDInsight"),
			ProductFamily: strPtr("Analytics"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "skuName", Value: strPtr(skuName)},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr("Consumption"),
		},
	}
}
