package schema

import (
	"github.com/tidwall/gjson"
)

type UsageData struct {
	Address    string
	Attributes map[string]gjson.Result
}

func NewUsageData(address string, attributes map[string]gjson.Result) *UsageData {
	return &UsageData{
		Address:    address,
		Attributes: attributes,
	}
}

func (u *UsageData) Get(key string) gjson.Result {
	return u.Attributes[key]
}
