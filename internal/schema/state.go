package schema

// State contains the existing, planned state of
// resources and the diff between them.
type State struct {
	ExistingState *ResourcesState
	PlannedState  *ResourcesState
	Diff          *ResourcesState
}
