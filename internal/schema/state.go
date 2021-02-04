package schema

// State contains the existing, planned state of
// resources and the diff between them.
type State struct {
	ExistingState *ResourcesState
	PlannedState  *ResourcesState
	Diff          *ResourcesState
}

// AllResources returns a pointer list of all resources of the state.
func (state *State) AllResources() []*Resource {
	var resources []*Resource
	resources = append(resources, state.ExistingState.Resources...)
	resources = append(resources, state.PlannedState.Resources...)
	resources = append(resources, state.Diff.Resources...)
	return resources
}

// CalculateTotalCosts will calculate and fill the total costs fields
// of State's ResourcesState. It must be called after calculating the costs of
// the resources.
func (state *State) CalculateTotalCosts() {
	state.ExistingState.calculateTotalCosts()
	state.PlannedState.calculateTotalCosts()
}
