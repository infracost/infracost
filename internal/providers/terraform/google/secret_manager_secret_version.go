package google

import (
	"github.com/infracost/infracost/internal/resources/google"
	"github.com/infracost/infracost/internal/schema"
)

func getSecretManagerSecretVersionRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "google_secret_manager_secret_version",
		CoreRFunc: newSecretManagerSecretVersion,
		ReferenceAttributes: []string{
			"secret",
		},
	}
}

func newSecretManagerSecretVersion(d *schema.ResourceData) schema.CoreResource {
	replicasCount := int64(1)

	secretReferences := d.References("secret")
	if len(secretReferences) > 0 {
		replicasCount = secretManagerSecretReplicasCount(secretReferences[0])
	}

	return &google.SecretManagerSecretVersion{
		Address:              d.Address,
		Region:               d.Get("region").String(),
		ReplicationLocations: replicasCount,
	}
}
