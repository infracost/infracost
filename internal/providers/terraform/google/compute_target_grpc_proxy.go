package google

import (
	"github.com/infracost/infracost/internal/resources/google"
	"github.com/infracost/infracost/internal/schema"
)

func getComputeTargetGrpcProxyRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "google_compute_target_grpc_proxy",
		RFunc: NewComputeTargetGrpcProxy,
		Notes: []string{"Price for additional forwarding rule is used"},
	}
}
func getComputeTargetHTTPProxyRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "google_compute_target_http_proxy",
		RFunc: NewComputeTargetGrpcProxy,
		Notes: []string{"Price for additional forwarding rule is used"},
	}
}
func getComputeTargetHTTPSProxyRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "google_compute_target_https_proxy",
		RFunc: NewComputeTargetGrpcProxy,
		Notes: []string{"Price for additional forwarding rule is used"},
	}
}
func getComputeTargetSslProxyRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "google_compute_target_ssl_proxy",
		RFunc: NewComputeTargetGrpcProxy,
		Notes: []string{"Price for additional forwarding rule is used"},
	}
}
func getComputeTargetTCPProxyRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "google_compute_target_tcp_proxy",
		RFunc: NewComputeTargetGrpcProxy,
		Notes: []string{"Price for additional forwarding rule is used"},
	}
}
func getComputeRegionTargetHTTPProxyRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "google_compute_region_target_http_proxy",
		RFunc: NewComputeTargetGrpcProxy,
		Notes: []string{"Price for additional forwarding rule is used"},
	}
}
func getComputeRegionTargetHTTPSProxyRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "google_compute_region_target_https_proxy",
		RFunc: NewComputeTargetGrpcProxy,
		Notes: []string{"Price for additional forwarding rule is used"},
	}
}

func NewComputeTargetGrpcProxy(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	r := &google.ComputeTargetGrpcProxy{Address: d.Address, Region: d.Get("region").String()}
	r.PopulateUsage(u)
	return r.BuildResource()
}
