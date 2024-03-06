package google

import (
	"github.com/infracost/infracost/internal/resources/google"
	"github.com/infracost/infracost/internal/schema"
)

func getMonitoringItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "google_monitoring_metric_descriptor",
		CoreRFunc: NewMonitoringMetricDescriptor,
	}
}

func NewMonitoringMetricDescriptor(d *schema.ResourceData) schema.CoreResource {
	r := &google.MonitoringMetricDescriptor{
		Address: d.Address,
	}

	return r
}
