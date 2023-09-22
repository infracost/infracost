package hcl

type VertexRoot struct{}

func (v *VertexRoot) ID() string {
	return "_root"
}

func (v *VertexRoot) Evaluator() *Evaluator {
	return nil
}

func (v *VertexRoot) References() []string {
	return []string{}
}

func (v *VertexRoot) Evaluate() error {
	return nil
}

func (v *VertexRoot) Expand() ([]*Block, error) {
	return []*Block{}, nil
}
