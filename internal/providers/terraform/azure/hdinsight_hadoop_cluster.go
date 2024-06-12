package azure

import (
	"fmt"
	"strings"

	"github.com/shopspring/decimal"
	"github.com/tidwall/gjson"

	"github.com/infracost/infracost/internal/schema"
)

func GetAzureRMHDInsightHadoopClusterRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "azurerm_hdinsight_hadoop_cluster", //nolint:misspell
		RFunc: NewAzureRMHDInsightHadoopCluster,
	}
}

func NewAzureRMHDInsightHadoopCluster(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	region := d.Region

	costComponents := []*schema.CostComponent{}

	headNodeVM := d.Get("roles.0.head_node.0.vm_size").String()
	workerNodeVM := d.Get("roles.0.worker_node.0.vm_size").String()
	var workerInstances int64
	if d.Get("roles.0.worker_node.0.target_instance_count").Type != gjson.Null {
		workerInstances = d.Get("roles.0.worker_node.0.target_instance_count").Int()
	}
	zookeeperNodeVM := d.Get("roles.0.zookeeper_node.0.vm_size").String()

	costComponents = append(costComponents, hdInsightVMCostComponent(region, "Head", headNodeVM, 2))
	costComponents = append(costComponents, hdInsightVMCostComponent(region, "Worker", workerNodeVM, workerInstances))
	costComponents = append(costComponents, hdInsightVMCostComponent(region, "Zookeeper", zookeeperNodeVM, 3))

	var edgeNodeVM string
	if d.Get("roles.0.edge_node.0.vm_size").Type != gjson.Null {
		edgeNodeVM = d.Get("roles.0.edge_node.0.vm_size").String()
		var workerInstances int64
		if d.Get("roles.0.edge_node.0.target_instance_count").Type != gjson.Null {
			workerInstances = d.Get("roles.0.edge_node.0.target_instance_count").Int()
		}
		costComponents = append(costComponents, hdInsightVMCostComponent(region, "Edge", edgeNodeVM, workerInstances))
	}

	return &schema.Resource{
		Name:           d.Address,
		CostComponents: costComponents,
	}
}

func hdInsightVMCostComponent(region, node, instanceType string, instances int64) *schema.CostComponent {
	t := strings.Split(instanceType, "_")
	dSeries := []string{"D1", "D2", "D3", "D4", "D5"}
	aSeries := []string{"A5", "A6", "A7", "A8", "A9", "A10", "A11"}
	if len(t) > 1 {
		if contains(dSeries, t[1]) {
			instanceType = fmt.Sprintf("%s_%s", t[0], t[1])
		}
	}
	if len(t) == 1 {
		if contains(aSeries, t[0]) {
			instanceType = fmt.Sprintf("Standard_%s", t[0])
		}
	}
	if len(t) == 3 {
		if strings.HasSuffix(strings.ToLower(instanceType), "v4") {
			instanceType = fmt.Sprintf("%s_%s %s", t[0], t[1], t[2])
		}
	}

	return &schema.CostComponent{
		Name:           fmt.Sprintf("%s node (%s)", node, instanceType),
		Unit:           "hours",
		UnitMultiplier: decimal.NewFromInt(1),
		HourlyQuantity: decimalPtr(decimal.NewFromInt(instances)),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("azure"),
			Region:        strPtr(region),
			Service:       strPtr("HDInsight"),
			ProductFamily: strPtr("Analytics"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "armSkuName", ValueRegex: strPtr(fmt.Sprintf("/^%s$/i", instanceType))},
			},
		},
		PriceFilter: &schema.PriceFilter{
			PurchaseOption: strPtr("Consumption"),
		},
	}
}
