package schema

import (
	"encoding/json"

	"github.com/shopspring/decimal"
)

type Suggestion struct {
	ID                 string           `json:"id"`
	Title              string           `json:"title"`
	Description        string           `json:"description"`
	ResourceType       string           `json:"resourceType"`
	ResourceAttributes json.RawMessage  `json:"resourceAttributes"`
	Address            string           `json:"address"`
	Suggested          string           `json:"suggested"`
	NoCost             bool             `json:"no_cost"`
	Cost               *decimal.Decimal `json:"cost"`
}

type Suggestions []Suggestion

func (s Suggestions) Len() int {
	return len(s)
}

func (s Suggestions) Less(i, j int) bool {
	iSug := s[i]
	jSug := s[j]

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

func (s Suggestions) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
