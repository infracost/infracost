package hcl

import (
	"fmt"

	"github.com/sirupsen/logrus"
)

type VertexLocal struct {
	logger    *logrus.Entry
	evaluator *Evaluator
	block     *Block
	attr      *Attribute
}

func (v *VertexLocal) ID() string {
	return fmt.Sprintf("%s%s", v.block.FullName(), v.attr.Name())
}

func (v *VertexLocal) ModuleAddress() string {
	return v.block.ModuleAddress()
}

func (v *VertexLocal) Evaluator() *Evaluator {
	return v.evaluator
}

func (v *VertexLocal) References() []VertexReference {
	return referencesForAttribute(v.block, v.attr)
}

func (v *VertexLocal) Evaluate() error {
	v.logger.Debugf("adding attribute %s to the evaluation context", v.ID())

	key := fmt.Sprintf("local.%s", v.attr.Name())
	v.evaluator.ctx.SetByDot(v.attr.Value(), key)

	return nil
}

func (v *VertexLocal) Expand() ([]*Block, error) {
	return []*Block{}, nil
}
