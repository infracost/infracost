package google

import (
	"github.com/infracost/infracost/internal/resources/google"
	"github.com/infracost/infracost/internal/schema"
)

func getBigQueryTableRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "google_bigquery_table",
		RFunc: NewBigQueryTable,
	}
}

func NewBigQueryTable(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	r := &google.BigQueryTable{
		Address: d.Address,
		Region:  d.Get("region").String(),
	}

	r.PopulateUsage(u)
	return r.BuildResource()
}
