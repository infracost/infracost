package modules

import (
	"sync"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
)

type SharedHCLParser struct {
	parser *hclparse.Parser
	mu     *sync.Mutex
}

func NewSharedHCLParser() *SharedHCLParser {
	return &SharedHCLParser{
		parser: hclparse.NewParser(),
		mu:     &sync.Mutex{},
	}
}

func (p *SharedHCLParser) ParseHCLFile(filename string) (*hcl.File, hcl.Diagnostics) {
	p.mu.Lock()
	defer p.mu.Unlock()

	return p.parser.ParseHCLFile(filename)
}

func (p *SharedHCLParser) ParseJSONFile(filename string) (*hcl.File, hcl.Diagnostics) {
	p.mu.Lock()
	defer p.mu.Unlock()

	return p.parser.ParseJSONFile(filename)
}
