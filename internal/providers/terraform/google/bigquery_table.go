package google

import (
	"github.com/infracost/infracost/internal/resources/google"
	"github.com/infracost/infracost/internal/schema"
)

func getBigqueryTableRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "google_bigquery_table",
		RFunc: NewBigqueryTable,
	}
}
func NewBigqueryTable(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	r := &google.BigqueryTable{Address: strPtr(d.Address), Region: strPtr(d.Get("region").String())}
	r.PopulateUsage(u)
	return r.BuildResource()
}
