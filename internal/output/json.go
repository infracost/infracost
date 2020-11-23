package output

import (
	"encoding/json"
)

func ToJSON(out Root) ([]byte, error) {
	return json.Marshal(out)
}
