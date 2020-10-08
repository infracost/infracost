package schema

import (
	"github.com/urfave/cli/v2"
)

type Provider interface {
	ProcessArgs(*cli.Context) error
	LoadResources() ([]*Resource, error)
}
