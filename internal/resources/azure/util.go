package azure

import (
	"fmt"
	"slices"
	"strings"

	"github.com/shopspring/decimal"

	"github.com/infracost/infracost/internal/schema"
)

const (
	vendorName = "azure"
)

var (
	priceFilterConsumption = &schema.PriceFilter{
		PurchaseOption: strPtr("Consumption"),
	}
	priceFilterDevTestConsumption = &schema.PriceFilter{
		PurchaseOption: strPtr("DevTestConsumption"),
	}
)

func strPtr(s string) *string {
	return &s
}

func decimalPtr(d decimal.Decimal) *decimal.Decimal {
	return &d
}

func int32Ptr(i int32) *int32 {
	return &i
}

func intPtrToDecimalPtr(i *int64) *decimal.Decimal {
	if i == nil {
		return nil
	}
	return decimalPtr(decimal.NewFromInt(*i))
}

func floatPtrToDecimalPtr(f *float64) *decimal.Decimal {
	if f == nil {
		return nil
	}
	return decimalPtr(decimal.NewFromFloat(*f))
}

func contains(a []string, x string) bool {
	return slices.Contains(a, x)
}

func containsInt64(arr []int64, val int64) bool {
	return slices.Contains(arr, val)
}

func regexPtr(regex string) *string {
	return strPtr(fmt.Sprintf("/%s/i", regex))
}

func convertRegion(region string) string {
	if strings.Contains(strings.ToLower(region), "usgov") {
		return "US Gov"
	} else if strings.Contains(strings.ToLower(region), "china") {
		return "Сhina"
	} else {
		return "Global"
	}
}
