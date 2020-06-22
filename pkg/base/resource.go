package base

import "github.com/shopspring/decimal"

type PriceComponent interface {
	Name() string
	Resource() Resource
	Filters() []Filter
	SetPrice(price decimal.Decimal)
	HourlyCost() decimal.Decimal
	SkipQuery() bool
}

type Resource interface {
	Address() string
	SubResources() []Resource
	PriceComponents() []PriceComponent
	References() map[string]Resource
	AddReference(name string, resource Resource)
	HasCost() bool
}

func GetPriceComponent(resource Resource, name string) PriceComponent {
	for _, p := range resource.PriceComponents() {
		if p.Name() == name {
			return p
		}
	}
	return nil
}
