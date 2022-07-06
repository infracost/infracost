package ibm

import (
	"github.com/infracost/infracost/internal/resources/ibm"
	"github.com/infracost/infracost/internal/schema"
	"github.com/tidwall/gjson"
)

func getIbmIsVpcRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "ibm_is_vpc",
		RFunc: newIbmIsVpc,
	}
}

func newIbmIsVpc(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	region := d.Get("region").String()
	r := &ibm.IbmIsVpc{
		Address: d.Address,
		Region:  region,
	}
	r.PopulateUsage(u)

	metadata := make(map[string]gjson.Result)
	properties := gjson.Result{
		Type: gjson.JSON,
		Raw: `{
			"serviceId": "is.vpc",
			"childResources": ["ibm_is_flow_log"]
		}`,
	}

	metadata["catalog"] = properties
	d.Metadata = metadata

	return r.BuildResource()
}
