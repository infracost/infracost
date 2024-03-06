package google

import (
	"github.com/infracost/infracost/internal/resources/google"
	"github.com/infracost/infracost/internal/schema"
)

func getBigQueryTableRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "google_bigquery_table",
		CoreRFunc: NewBigQueryTable,
	}
}

func NewBigQueryTable(d *schema.ResourceData) schema.CoreResource {
	r := &google.BigQueryTable{
		Address: d.Address,
		Region:  d.Get("region").String(),
	}

	return r
}
