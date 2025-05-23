package google

import (
	"github.com/infracost/infracost/internal/resources/google"
	"github.com/infracost/infracost/internal/schema"
)

func getKMSCryptoKeyRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "google_kms_crypto_key",
		RFunc: NewKMSCryptoKey,
	}
}
func NewKMSCryptoKey(d *schema.ResourceData, u *schema.UsageData) *schema.Resource {
	r := &google.KMSCryptoKey{Address: d.Address, Region: d.Get("region").String(), Algorithm: d.Get("versionTemplate.0.algorithm").String(), ProtectionLevel: d.Get("versionTemplate.0.protectionLevel").String(), RotationPeriod: d.Get("rotationPeriod").String(), VersionTemplate: d.Get("versionTemplate").String()}
	r.PopulateUsage(u)
	return r.BuildResource()
}
