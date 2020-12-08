package aws

import (
	"github.com/infracost/infracost/internal/schema"
	"github.com/shopspring/decimal"
)

func GetNewKMSKeyRegistryItem() *schema.RegistryItem {
	return &schema.RegistryItem{
		Name:  "aws_kms_key",
		RFunc: NewKMSKey,
	}
}

func NewKMSKey(d *schema.ResourceData, u *schema.ResourceData) *schema.Resource {

	region := d.Get("region").String()
	spec := d.Get("customer_master_key_spec").String()

	costComponents := []*schema.CostComponent{
		{
			Name:            "Customer master key",
			Unit:            "months",
			UnitMultiplier:  1,
			MonthlyQuantity: decimalPtr(decimal.NewFromInt(1)),
			ProductFilter: &schema.ProductFilter{
				VendorName:    strPtr("aws"),
				Region:        strPtr(region),
				Service:       strPtr("awskms"),
				ProductFamily: strPtr("Encryption Key"),
				AttributeFilters: []*schema.AttributeFilter{
					{Key: "locationType", Value: strPtr("AWS Region")},
					{Key: "usagetype", ValueRegex: strPtr("/KMS-Keys/")},
				},
			},
		},
	}

	costComponents = appendRequestComponentsForSpec(costComponents, spec, region)

	return &schema.Resource{
		Name:           d.Address,
		CostComponents: costComponents,
	}
}

func appendRequestComponentsForSpec(costComponents []*schema.CostComponent, spec string, region string) []*schema.CostComponent {

	switch spec {
	case "RSA_2048":
		costComponents = append(costComponents, requestPriceComponent("Requests (RSA 2048)", region, "Asymmetric-RSA_2048"))
		return costComponents
	case
		"RSA_3072",
		"RSA_4096",
		"ECC_NIST_P256",
		"ECC_NIST_P384",
		"ECC_NIST_P521",
		"ECC_SECG_P256K1":
		costComponents = append(costComponents, requestPriceComponent("Requests (asymmetric)", region, "Asymmetric"))
		return costComponents
	}

	costComponents = append(costComponents, requestPriceComponent("Requests", region, "All"))
	costComponents = append(costComponents, requestPriceComponent("ECC GenerateDataKeyPair requests", region, "GenerateDatakeyPair-ECC"))
	costComponents = append(costComponents, requestPriceComponent("RSA GenerateDataKeyPair requests", region, "GenerateDatakeyPair-ECC"))
	return costComponents
}

func requestPriceComponent(name string, region string, group string) *schema.CostComponent {
	return &schema.CostComponent{
		Name:           name,
		Unit:           "requests",
		UnitMultiplier: 10000,
		ProductFilter: &schema.ProductFilter{
			VendorName: strPtr("aws"),
			Region:     strPtr(region),
			Service:    strPtr("awskms"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "group", ValueRegex: strPtr("/" + group + "/")},
				{Key: "locationType", Value: strPtr("AWS Region")},
				{Key: "usagetype", ValueRegex: strPtr("/KMS-Requests/")},
			},
		},
	}
}
