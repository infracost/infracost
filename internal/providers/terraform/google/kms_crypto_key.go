package google

import (
	"github.com/infracost/infracost/internal/resources/google"
	"github.com/infracost/infracost/internal/schema"
)

func getKMSCryptoKeyRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "google_kms_crypto_key",
		RFunc: NewKMSCryptoKey,
	}
}
func NewKMSCryptoKey(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	r := &google.KMSCryptoKey{Address: d.Address, Region: d.Get("region").String(), Algorithm: d.Get("version_template.0.algorithm").String(), ProtectionLevel: d.Get("version_template.0.protection_level").String(), RotationPeriod: d.Get("rotation_period").String(), VersionTemplate: d.Get("version_template").String()}
	r.PopulateUsage(u)
	return r.BuildResource()
}
