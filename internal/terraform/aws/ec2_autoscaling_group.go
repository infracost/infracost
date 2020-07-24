package aws

import (
	"fmt"
	"infracost/pkg/base"
	"math"
	"strconv"
)

type Ec2AutoscalingGroupResource struct {
	*base.BaseResource
	region string
}

func NewEc2AutoscalingGroup(address string, region string, rawValues map[string]interface{}) base.Resource {
	r := &Ec2AutoscalingGroupResource{
		base.NewBaseResource(address, rawValues, true),
		region,
	}
	return r
}

func (r *Ec2AutoscalingGroupResource) AddReference(name string, resource base.Resource) {
	r.BaseResource.AddReference(name, resource)

	capacity := int(r.RawValues()["desired_capacity"].(float64))
	address := fmt.Sprintf("%s.%s", r.Address(), resource.Address())

	if name == "launch_configuration" || name == "launch_configuration_id" {
		launchConfiguration := NewEc2LaunchConfiguration(address, r.region, resource.RawValues(), true)
		launchConfiguration.SetResourceCount(capacity)
		r.AddSubResource(launchConfiguration)

	} else if name == "launch_template" || name == "launch_template_id" {
		var overrides []interface{}
		overridesVal := base.ToGJSON(r.RawValues()).Get("mixed_instances_policy.0.launch_template.0.override").Value()
		if overridesVal != nil {
			overrides = overridesVal.([]interface{})
		}

		if len(overrides) > 0 {
			// Just use the first override for now, since that will be the highest priority
			override := overrides[0].(map[string]interface{})
			instanceType := override["instance_type"].(string)
			weightedCapacity, _ := strconv.Atoi(override["weighted_capacity"].(string))
			count := int(math.Ceil(float64(capacity) / float64(weightedCapacity)))

			rawValues := make(map[string]interface{})
			for k, v := range resource.RawValues() {
				rawValues[k] = v
			}
			rawValues["instance_type"] = instanceType
			launchTemplate := NewEc2LaunchTemplate(address, r.region, rawValues, true)
			launchTemplate.SetResourceCount(count)
			r.AddSubResource(launchTemplate)

		} else {
			launchTemplate := NewEc2LaunchTemplate(address, r.region, resource.RawValues(), true)
			launchTemplate.SetResourceCount(capacity)
			r.AddSubResource(launchTemplate)
		}
	}
}
