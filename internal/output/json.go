package output

import (
	"encoding/json"

	"github.com/infracost/infracost/internal/config"
)

func ToJSON(ctx *config.RunContext, out Root, opts Options) ([]byte, error) {
	return json.Marshal(out)
}
