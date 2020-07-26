package aws

import (
	"fmt"
	"infracost/pkg/resource"
	"math"
	"strconv"
)

type Ec2AutoscalingGroupResource struct {
	*resource.BaseResource
	region string
}

func NewEc2AutoscalingGroup(address string, region string, rawValues map[string]interface{}) resource.Resource {
	r := &Ec2AutoscalingGroupResource{
		resource.NewBaseResource(address, rawValues, true),
		region,
	}
	return r
}

func (r *Ec2AutoscalingGroupResource) AddReference(name string, refResource resource.Resource) {
	r.BaseResource.AddReference(name, r)

	capacity := int(r.RawValues()["desired_capacity"].(float64))
	address := fmt.Sprintf("%s.%s", r.Address(), refResource.Address())

	if name == "launch_configuration" || name == "launch_configuration_id" {
		launchConfiguration := NewEc2LaunchConfiguration(address, r.region, refResource.RawValues(), true)
		launchConfiguration.SetResourceCount(capacity)
		r.AddSubResource(launchConfiguration)

	} else if name == "launch_template" || name == "launch_template_id" {
		var overrides []interface{}
		overridesVal := resource.ToGJSON(r.RawValues()).Get("mixed_instances_policy.0.launch_template.0.override").Value()
		if overridesVal != nil {
			overrides = overridesVal.([]interface{})
		}

		if len(overrides) > 0 {
			// Just use the first override for now, since that will be the highest priority
			override := overrides[0].(map[string]interface{})
			instanceType := override["instance_type"].(string)
			weightedCapacity := 1
			if override["weighted_capacity"] != nil {
				weightedCapacity, _ = strconv.Atoi(override["weighted_capacity"].(string))
			}
			count := int(math.Ceil(float64(capacity) / float64(weightedCapacity)))

			rawValues := make(map[string]interface{})
			for k, v := range refResource.RawValues() {
				rawValues[k] = v
			}
			rawValues["instance_type"] = instanceType
			launchTemplate := NewEc2LaunchTemplate(address, r.region, rawValues, true)
			launchTemplate.SetResourceCount(count)
			r.AddSubResource(launchTemplate)

		} else {
			launchTemplate := NewEc2LaunchTemplate(address, r.region, refResource.RawValues(), true)
			launchTemplate.SetResourceCount(capacity)
			r.AddSubResource(launchTemplate)
		}
	}
}
