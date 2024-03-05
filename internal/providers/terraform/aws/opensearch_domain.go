package aws

import (
	"github.com/infracost/infracost/internal/schema"
)

func getOpensearchDomainRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:      "aws_opensearch_domain",
		CoreRFunc: newSearchDomain,
	}
}
