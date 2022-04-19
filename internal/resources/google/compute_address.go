package google

import (
	"fmt"

	"github.com/infracost/infracost/internal/resources"
	"github.com/infracost/infracost/internal/schema"
	log "github.com/sirupsen/logrus"

	"strings"

	"github.com/shopspring/decimal"
)

type ComputeAddress struct {
	Address     string
	Region      string
	AddressType string

	// usage args
	AddressUsageType *string `infracost_usage:"address_type"`
}

var ComputeAddressUsageSchema = []*schema.UsageItem{
	{Key: "address_type", ValueType: schema.String, DefaultValue: ""},
}

func (r *ComputeAddress) PopulateUsage(u *schema.UsageData) {
	resources.PopulateArgsWithUsage(r, u)
}

func (r *ComputeAddress) BuildResource() *schema.Resource {
	addressType := r.AddressType
	if strings.ToLower(addressType) == "internal" {
		return &schema.Resource{
			Name:      r.Address,
			NoPrice:   true,
			IsSkipped: true, UsageSchema: ComputeAddressUsageSchema,
		}
	}

	costComponents := []*schema.CostComponent{}

	usageType, err := r.validateAddressUsageType()
	if err != "" {
		log.Warnf(err)
	}

	switch usageType {
	case "standard_vm":
		costComponents = append(costComponents, r.standardVMComputeAddress(true))
	case "preemptible_vm":
		costComponents = append(costComponents, r.preemptibleVMComputeAddress(true))
	case "unused":
		costComponents = append(costComponents, r.unusedVMComputeAddress(true))
	default:
		costComponents = append(costComponents,
			r.standardVMComputeAddress(false),
			r.preemptibleVMComputeAddress(false),
			r.unusedVMComputeAddress(false),
		)
	}

	return &schema.Resource{
		Name:           r.Address,
		CostComponents: costComponents,
		UsageSchema:    ComputeAddressUsageSchema,
	}
}

func (r *ComputeAddress) standardVMComputeAddress(used bool) *schema.CostComponent {
	usedBy := ""
	if !used {
		usedBy = "if used by "
	}

	return &schema.CostComponent{
		Name:           fmt.Sprintf("IP address (%sstandard VM)", usedBy),
		Unit:           "hours",
		UnitMultiplier: decimal.NewFromInt(1),
		HourlyQuantity: decimalPtr(decimal.NewFromInt(1)),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("gcp"),
			Region:        strPtr("global"),
			Service:       strPtr("Compute Engine"),
			ProductFamily: strPtr("Network"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "description", Value: strPtr("External IP Charge on a Standard VM")},
			},
		},
		PriceFilter: &schema.PriceFilter{
			StartUsageAmount: strPtr("744"),
		},
	}
}

func (r *ComputeAddress) preemptibleVMComputeAddress(used bool) *schema.CostComponent {
	usedBy := ""
	if !used {
		usedBy = "if used by "
	}

	return &schema.CostComponent{
		Name:           fmt.Sprintf("IP address (%spreemptible VM)", usedBy),
		Unit:           "hours",
		UnitMultiplier: decimal.NewFromInt(1),
		HourlyQuantity: decimalPtr(decimal.NewFromInt(1)),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("gcp"),
			Region:        strPtr("global"),
			Service:       strPtr("Compute Engine"),
			ProductFamily: strPtr("Network"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "description", Value: strPtr("External IP Charge on a Spot Preemptible VM")},
			},
		},
		PriceFilter: &schema.PriceFilter{
			EndUsageAmount: strPtr(""),
		},
	}
}

func (r *ComputeAddress) unusedVMComputeAddress(used bool) *schema.CostComponent {
	usedBy := ""
	if !used {
		usedBy = "if "
	}

	return &schema.CostComponent{
		Name:           fmt.Sprintf("IP address (%sunused)", usedBy),
		Unit:           "hours",
		UnitMultiplier: decimal.NewFromInt(1),
		HourlyQuantity: decimalPtr(decimal.NewFromInt(1)),
		ProductFilter: &schema.ProductFilter{
			VendorName:    strPtr("gcp"),
			Region:        strPtr(r.Region),
			Service:       strPtr("Compute Engine"),
			ProductFamily: strPtr("Network"),
			AttributeFilters: []*schema.AttributeFilter{
				{Key: "description", Value: strPtr("Static Ip Charge")},
			},
		},
		PriceFilter: &schema.PriceFilter{
			EndUsageAmount: strPtr(""),
		},
	}
}

func (r *ComputeAddress) validateAddressUsageType() (string, string) {
	validTypes := []string{"standard_vm", "preemptible_vm", "unused"}

	usageType := ""

	if r.AddressUsageType == nil {
		return usageType, ""
	}

	usageType = strings.ToLower(*r.AddressUsageType)

	if !contains(validTypes, usageType) {
		return "", fmt.Sprintf("Invalid address_type, ignoring. Expected: standard_vm, preemptible_vm, unused. Got: %s", *r.AddressUsageType)
	}

	return usageType, ""
}
