package schema

// Project contains the existing, planned state of
// resources and the diff between them.
type Project struct {
	Name          string
	PastResources []*Resource
	Resources     []*Resource
	Diff          []*Resource
	HasDiff       bool
}

func NewProject(name string) *Project {
	return &Project{
		Name:    name,
		HasDiff: true,
	}
}

// AllResources returns a pointer list of all resources of the state.
func (p *Project) AllResources() []*Resource {
	var resources []*Resource
	resources = append(resources, p.PastResources...)
	resources = append(resources, p.Resources...)
	return resources
}

// CalculateDiff calculates the diff of past and current resources
func (p *Project) CalculateDiff() {
	if p.HasDiff {
		p.Diff = calculateDiff(p.PastResources, p.Resources)
	}
}

// AllProjectResources returns the resources for all projects
func AllProjectResources(projects []*Project) []*Resource {
	resources := make([]*Resource, 0)

	for _, p := range projects {
		resources = append(resources, p.Resources...)
	}

	return resources
}
