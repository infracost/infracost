package schema

type Resource struct {
	Name           string
	SubResources   []*Resource
	CostComponents []*CostComponent
}
