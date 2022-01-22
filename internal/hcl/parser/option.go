package parser

import (
	"fmt"
	"os"
	"path/filepath"
)

type Option func(p *Parser)

func OptionDoNotSearchTfFiles() Option {
	return func(p *Parser) {
		p.stopOnFirstTf = false
	}
}

func OptionWithTFVarsPaths(paths []string) Option {
	return func(p *Parser) {
		p.tfvarsPaths = paths
	}
}

func OptionStopOnHCLError() Option {
	return func(p *Parser) {
		p.stopOnHCLError = true
	}
}

func OptionWithWorkspaceName(workspaceName string) Option {
	return func(p *Parser) {
		p.workspaceName = workspaceName
	}
}

func TfVarsToOption(varsPaths ...string) (Option, error) {
	for _, p := range varsPaths {
		tfvp, _ := filepath.Abs(p)
		_, err := os.Stat(tfvp)
		if err != nil {
			return nil, fmt.Errorf("passed tfvar file does not exist")
		}
	}

	return OptionWithTFVarsPaths(varsPaths), nil
}
