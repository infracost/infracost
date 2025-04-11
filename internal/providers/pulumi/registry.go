package pulumi

import (
	"github.com/infracost/infracost/internal/schema"
)

// ResourceRegistry contains a mapping between pulumi resource types and infracost resources
var ResourceRegistry []*schema.RegistryItem