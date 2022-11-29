package schema

import (
	"encoding/json"

	"github.com/shopspring/decimal"
)

type Policy struct {
	ID                 string           `json:"id"`
	Title              string           `json:"title"`
	Description        string           `json:"description"`
	ResourceType       string           `json:"resource_type"`
	ResourceAttributes json.RawMessage  `json:"resource_attributes"`
	Address            string           `json:"address"`
	Suggested          string           `json:"suggested"`
	NoCost             bool             `json:"no_cost"`
	Cost               *decimal.Decimal `json:"cost"`
}

type Policies []Policy

func (r Policies) Len() int {
	return len(r)
}

func (r Policies) Less(i, j int) bool {
	iSug := r[i]
	jSug := r[j]

	if iSug.Cost == nil && jSug.Cost == nil {
		return iSug.Address < jSug.Address
	}

	if iSug.Cost == nil {
		return false
	}

	if jSug.Cost == nil {
		return true
	}

	if iSug.Cost.Equal(*jSug.Cost) {
		return iSug.Address < jSug.Address
	}

	return iSug.Cost.GreaterThan(*jSug.Cost)
}

func (r Policies) Swap(i, j int) {
	r[i], r[j] = r[j], r[i]
}
