package output

import (
	"encoding/json"
)

func ToJSON(out Root, opts Options) ([]byte, error) {
	return json.Marshal(out)
}
