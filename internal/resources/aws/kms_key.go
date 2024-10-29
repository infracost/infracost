package aws

import (
	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"

	"github.com/shopspring/decimal"
)

type KMSKey struct {
	Address               string
	Region                string
	CustomerMasterKeySpec string
}

func (r *KMSKey) CoreType() string {
	return "KMSKey"
}

func (r *KMSKey) UsageSchema() []*schema.UsageItem {
	return []*schema.UsageItem{}
}

func (r *KMSKey) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *KMSKey) BuildResource() *schema.Resource {
	costComponents := []*schema.CostComponent{
		r.customerMasterKeyCostComponent(),
	}

	costComponents = append(costComponents, r.requestsCostComponents()...)

	return &schema.Resource{
		Name:           r.Address,
		CostComponents: costComponents,
		UsageSchema:    r.UsageSchema(),
	}
}

func (r *KMSKey) customerMasterKeyCostComponent() *schema.CostComponent {
	return &schema.CostComponent{
		Name:            "Customer master key",
		Unit:            "months",
		UnitMultiplier:  decimal.NewFromInt(1),
		MonthlyQuantity: decimalPtr(decimal.NewFromInt(1)),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("aws"),
			Region:        strPtr(r.Region),
			Service:       strPtr("awskms"),
			ProductFamily: strPtr("Encryption Key"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "usagetype", ValueRegex: strPtr("/KMS-Keys/")},
			},
		},
	}
}

func (r *KMSKey) requestsCostComponents() []*schema.CostComponent {
	switch r.CustomerMasterKeySpec {
	case "RSA_2048":
		return []*schema.CostComponent{
			r.requestsCostComponent("Requests (RSA 2048)", "/KMS-Requests-Asymmetric-RSA_2048/"),
		}
	case
		"RSA_3072",
		"RSA_4096",
		"ECC_NIST_P256",
		"ECC_NIST_P384",
		"ECC_NIST_P521",
		"ECC_SECG_P256K1":
		return []*schema.CostComponent{
			r.requestsCostComponent("Requests (asymmetric)", "/KMS-Requests-Asymmetric$/"),
		}
	}

	return []*schema.CostComponent{
		r.requestsCostComponent("Requests", "/KMS-Requests$/"),
		r.requestsCostComponent("ECC GenerateDataKeyPair requests", "/KMS-Requests-GenerateDatakeyPair-ECC/"),
		r.requestsCostComponent("RSA GenerateDataKeyPair requests", "/KMS-Requests-GenerateDatakeyPair-ECC/"),
	}
}

func (r *KMSKey) requestsCostComponent(name string, usagetype string) *schema.CostComponent {
	return &schema.CostComponent{
		Name:           name,
		Unit:           "10k requests",
		UnitMultiplier: decimal.NewFromInt(10000),
		ProductFilter: &schema.ProductFilter{
			VendorName: strPtr("aws"),
			Region:     strPtr(r.Region),
			Service:    strPtr("awskms"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "usagetype", ValueRegex: strPtr(usagetype)},
			},
		},
	}
}
