package k8s

import (
	"fmt"

	"github.com/shopspring/decimal"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/infracost/infracost/internal/schema"
)

func NewDeploymentResource(object runtime.Object) (*schema.Resource, error) {
	deployment := object.(*appsv1.Deployment)
	var replicas float64 = 1
	if deployment.Spec.Replicas != nil {
		replicas = float64(*deployment.Spec.Replicas)
	}
	namespace := "default"
	if deployment.Namespace != "" {
		namespace = deployment.Namespace
	}

	tags := map[string]string{
		"namespace": namespace,
		"name":      deployment.Name,
	}

	for k, v := range deployment.Labels {
		tags[fmt.Sprintf("labels.%s", k)] = v
	}

	name := fmt.Sprintf("deployment/%s", deployment.Name)
	extra := ""
	if replicas > 1 {
		extra = fmt.Sprintf("replicas: %d, ", int(replicas))
	}

	extra += fmt.Sprintf("namespace: %s", namespace)
	name = fmt.Sprintf("%s (%s)", name, extra)

	parent := &schema.Resource{
		Name: name,
		Tags: tags,
	}

	for _, container := range deployment.Spec.Template.Spec.Containers {
		parent.SubResources = append(parent.SubResources, &schema.Resource{
			Name: fmt.Sprintf("container/%s", container.Name),
			CostComponents: []*schema.CostComponent{
				cpuCostComponent(container, replicas),
				ramCostComponent(container, replicas),
			},
		})
	}

	return parent, nil
}

func ramCostComponent(container v1.Container, replicas float64) *schema.CostComponent {
	var ram *decimal.Decimal

	if container.Resources.Requests != nil {
		ram = schema.DecimalPtr(decimal.NewFromFloat(float64(container.Resources.Requests.Memory().Value()) / (1024 * 1024) * replicas))
	} else if container.Resources.Limits != nil {
		ram = schema.DecimalPtr(decimal.NewFromFloat(float64(container.Resources.Limits.Memory().Value()) / (1024 * 1024) * replicas))
	}

	ramCostComponent := &schema.CostComponent{
		Name:            "RAM",
		Unit:            "MiB",
		UnitMultiplier:  schema.DecimalOne,
		MonthlyQuantity: ram,
		ProductFilter: &schema.ProductFilter{
			Sku: schema.StringPtr("ram"),
		},
	}
	return ramCostComponent
}

func cpuCostComponent(container v1.Container, replicas float64) *schema.CostComponent {
	var cpu *decimal.Decimal

	if container.Resources.Requests != nil {
		cpu = schema.DecimalPtr(decimal.NewFromFloat(float64(container.Resources.Requests.Cpu().MilliValue()) * replicas))
	} else if container.Resources.Limits != nil {
		cpu = schema.DecimalPtr(decimal.NewFromFloat(float64(container.Resources.Limits.Cpu().MilliValue()) * replicas))
	}

	unit := "millicores"
	multiplier := schema.DecimalOne
	coreQty := decimal.NewFromInt(1000)

	if cpu != nil && cpu.GreaterThanOrEqual(coreQty) {
		unit = "cores"
		multiplier = coreQty
	}

	return &schema.CostComponent{
		Name:            "CPU",
		Unit:            unit,
		UnitMultiplier:  multiplier,
		MonthlyQuantity: cpu,
		ProductFilter: &schema.ProductFilter{
			Sku: schema.StringPtr("cpu"),
		},
	}
}
