package google

import (
	"github.com/infracost/infracost/internal/resources/google"
	"github.com/infracost/infracost/internal/schema"
)

func getComputeTargetGRPCProxyRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "google_compute_target_grpc_proxy",
		CoreRFunc: NewComputeTargetGRPCProxy,
		Notes:     []string{"Price for additional forwarding rule is used"},
	}
}
func getComputeTargetHTTPProxyRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "google_compute_target_http_proxy",
		CoreRFunc: NewComputeTargetGRPCProxy,
		Notes:     []string{"Price for additional forwarding rule is used"},
	}
}
func getComputeTargetHTTPSProxyRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "google_compute_target_https_proxy",
		CoreRFunc: NewComputeTargetGRPCProxy,
		Notes:     []string{"Price for additional forwarding rule is used"},
	}
}
func getComputeTargetSSLProxyRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "google_compute_target_ssl_proxy",
		CoreRFunc: NewComputeTargetGRPCProxy,
		Notes:     []string{"Price for additional forwarding rule is used"},
	}
}
func getComputeTargetTCPProxyRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "google_compute_target_tcp_proxy",
		CoreRFunc: NewComputeTargetGRPCProxy,
		Notes:     []string{"Price for additional forwarding rule is used"},
	}
}
func getComputeRegionTargetHTTPProxyRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "google_compute_region_target_http_proxy",
		CoreRFunc: NewComputeTargetGRPCProxy,
		Notes:     []string{"Price for additional forwarding rule is used"},
	}
}
func getComputeRegionTargetHTTPSProxyRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "google_compute_region_target_https_proxy",
		CoreRFunc: NewComputeTargetGRPCProxy,
		Notes:     []string{"Price for additional forwarding rule is used"},
	}
}

func NewComputeTargetGRPCProxy(d *schema.ResourceData) schema.CoreResource {
	r := &google.ComputeTargetGRPCProxy{Address: d.Address, Region: d.Get("region").String()}
	return r
}
