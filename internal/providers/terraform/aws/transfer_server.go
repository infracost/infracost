package aws

import (
	"sort"

	"github.com/infracost/infracost/internal/resources/aws"
	"github.com/infracost/infracost/internal/schema"
)

func getTransferServerRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "aws_transfer_server",
		CoreRFunc: newTransferServer,
	}
}

func newTransferServer(d *schema.ResourceData) schema.CoreResource {
	region := d.Get("region").String()
	protocols := []string{}

	if d.Get("protocols").Exists() {
		for _, data := range d.Get("protocols").Array() {
			protocols = append(protocols, data.String())
		}

		sort.Strings(protocols)
	} else {
		defaultProtocol := "SFTP"
		protocols = append(protocols, defaultProtocol)
	}

	t := &aws.TransferServer{
		Address:   d.Address,
		Region:    region,
		Protocols: protocols,
	}

	return t
}
