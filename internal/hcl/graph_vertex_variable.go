package hcl

import (
	"fmt"

	"github.com/sirupsen/logrus"
)

type VertexVariable struct {
	logger    *logrus.Entry
	evaluator *Evaluator
	block     *Block
}

func (v *VertexVariable) ID() string {
	return v.block.FullName()
}

func (v *VertexVariable) ModuleAddress() string {
	return v.block.ModuleAddress()
}

func (v *VertexVariable) Evaluator() *Evaluator {
	return v.evaluator
}

func (v *VertexVariable) References() []VertexReference {
	return referencesForBlock(v.block)
}

func (v *VertexVariable) Evaluate() error {
	val, err := v.evaluator.evaluateVariable(v.block)
	if err != nil {
		return fmt.Errorf("could not evaluate variable %s: %w", v.ID(), err)
	}

	v.logger.Debugf("adding variable %s to the evaluation context", v.ID())
	key := fmt.Sprintf("var.%s", v.block.Label())
	v.evaluator.ctx.SetByDot(val, key)

	return nil
}

func (v *VertexVariable) Expand() ([]*Block, error) {
	return []*Block{v.block}, nil
}
