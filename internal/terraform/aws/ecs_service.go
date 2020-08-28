package aws

import (
	"fmt"
	"infracost/pkg/resource"
	"regexp"

	"github.com/shopspring/decimal"
)

type EcsServiceResource struct {
	*resource.BaseResource
	region string
}

func convertResourceString(rawValue string) decimal.Decimal {
	var quantity decimal.Decimal
	reg := regexp.MustCompile(`(?i)\s*(vcpu|gb)\s*`)
	if reg.MatchString(rawValue) {
		quantity, _ = decimal.NewFromString(reg.ReplaceAllString(rawValue, ""))
	} else {
		quantity, _ = decimal.NewFromString(rawValue)
		quantity = quantity.Div(decimal.NewFromInt(1024))
	}
	return quantity
}

func quantityFactory(quantity decimal.Decimal) func(resource resource.Resource) decimal.Decimal {
	return func(resource resource.Resource) decimal.Decimal {
		return quantity
	}
}

func (r *EcsServiceResource) AddReference(name string, refResource resource.Resource) {
	r.BaseResource.AddReference(name, r)

	count := 0
	countVal := r.RawValues()["desired_count"]
	if countVal != nil {
		count = int(countVal.(float64))
	}
	r.SetResourceCount(count)

	if r.RawValues()["launch_type"] != nil && r.RawValues()["launch_type"].(string) == "FARGATE" {
		if name == "task_definition" {
			memory := convertResourceString(refResource.RawValues()["memory"].(string))
			gbHoursProductFilter := &resource.ProductFilter{
				VendorName:    strPtr("aws"),
				Region:        strPtr(r.region),
				Service:       strPtr("AmazonECS"),
				ProductFamily: strPtr("Compute"),
				AttributeFilters: &[]resource.AttributeFilter{
					{Key: "usagetype", ValueRegex: strPtr("/Fargate-GB-Hours/")},
				},
			}
			gbHours := resource.NewBasePriceComponent("GB hours", r, "GB-hour", "hour", gbHoursProductFilter, nil)
			gbHours.SetQuantityMultiplierFunc(quantityFactory(memory))
			r.AddPriceComponent(gbHours)

			cpu := convertResourceString(refResource.RawValues()["cpu"].(string))
			cpuHoursProductFilter := &resource.ProductFilter{
				VendorName:    strPtr("aws"),
				Region:        strPtr(r.region),
				Service:       strPtr("AmazonECS"),
				ProductFamily: strPtr("Compute"),
				AttributeFilters: &[]resource.AttributeFilter{
					{Key: "usagetype", ValueRegex: strPtr("/Fargate-vCPU-Hours:perCPU/")},
				},
			}
			cpuHours := resource.NewBasePriceComponent("CPU hours", r, "CPU-hour", "hour", cpuHoursProductFilter, nil)
			cpuHours.SetQuantityMultiplierFunc(quantityFactory(cpu))
			r.AddPriceComponent(cpuHours)

			acceleratorRawValue := refResource.RawValues()["inference_accelerator"]
			if acceleratorRawValue != nil && len(acceleratorRawValue.([]interface{})) > 0 {
				deviceType := resource.ToGJSON(refResource.RawValues()).Get("inference_accelerator.0.device_type")
				acceleratorHoursProductFilter := &resource.ProductFilter{
					VendorName:    strPtr("aws"),
					Region:        strPtr(r.region),
					Service:       strPtr("AmazonEI"),
					ProductFamily: strPtr("Elastic Inference"),
					AttributeFilters: &[]resource.AttributeFilter{
						{Key: "usagetype", ValueRegex: strPtr(fmt.Sprintf("/%s/", deviceType))},
					},
				}
				acceleratorHours := resource.NewBasePriceComponent(fmt.Sprintf("Accelerator hours (%s)", deviceType), r, "hour", "hour", acceleratorHoursProductFilter, nil)
				r.AddPriceComponent(acceleratorHours)
			}
		}
	}
	// } else {
	//  TODO: support ECS-EC2 pricing
	// 	// productFamily = "Compute Metering" with usagetypes = ["/ECS-EC2-GB-Hours/", "/ECS-EC2-vCPU-Hours/"]
	// }
}

func NewEcsService(address string, region string, rawValues map[string]interface{}) resource.Resource {
	r := &EcsServiceResource{
		resource.NewBaseResource(address, rawValues, true),
		region,
	}

	return r
}
