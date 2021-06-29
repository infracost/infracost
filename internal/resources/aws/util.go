package aws

import (
	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
)

func strPtr(s string) *string {
	return &s
}

func decimalPtr(d decimal.Decimal) *decimal.Decimal {
	return &d
}

func setShouldSync(schema []*schema.UsageSchemaItem, keysToSkipSync []string) {
	skipSync := stringSet(keysToSkipSync)
	for _, schemaItem := range schema {
		schemaItem.ShouldSync = !skipSync[schemaItem.Key]
	}
}

func stringSet(s []string) map[string]bool {
	set := make(map[string]bool)
	for _, k := range s {
		set[k] = true
	}
	return set
}
