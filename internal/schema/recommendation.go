package schema

import (
	"encoding/json"

	"github.com/shopspring/decimal"
)

type Recommendation struct {
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

type Recommendations []Recommendation

func (r Recommendations) Len() int {
	return len(r)
}

func (r Recommendations) Less(i, j int) bool {
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

func (r Recommendations) Swap(i, j int) {
	r[i], r[j] = r[j], r[i]
}
