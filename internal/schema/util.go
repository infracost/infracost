package schema

import "github.com/shopspring/decimal"

func StringPtr(str string) *string {
	return &str
}

func DecimalPtr(dec decimal.Decimal) *decimal.Decimal {
	return &dec
}
