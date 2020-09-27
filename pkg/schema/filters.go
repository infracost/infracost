package schema

type ProductFilter struct {
	VendorName       *string            `json:"vendorName,omitempty"`
	Service          *string            `json:"service,omitempty"`
	ProductFamily    *string            `json:"productFamily,omitempty"`
	Region           *string            `json:"region,omitempty"`
	Sku              *string            `json:"sku,omitempty"`
	AttributeFilters []*AttributeFilter `json:"attributeFilters,omitempty"`
}

type PriceFilter struct {
	PurchaseOption     *string `json:"purchaseOption,omitempty"`
	Unit               *string `json:"unit,omitempty"`
	Description        *string `json:"description,omitempty"`
	DescriptionRegex   *string `json:"description_regex,omitempty"`
	StartUsageAmount   *string `json:"startUsageAmount,omitempty"`
	EndUsageAmount     *string `json:"endUsageAmount,omitempty"`
	TermLength         *string `json:"termLength,omitempty"`
	TermPurchaseOption *string `json:"termPurchaseOption,omitempty"`
	TermOfferingClass  *string `json:"termOfferingClass,omitempty"`
}

type AttributeFilter struct {
	Key        string  `json:"key"`
	Value      *string `json:"value,omitempty"`
	ValueRegex *string `json:"value_regex,omitempty"`
}
