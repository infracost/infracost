package schema

// ReferenceIDFunc is used to let references be built using non-standard IDs (anything other than d.Get("id").string)
type ReferenceIDFunc func(d *ResourceData) []string

type RegistryItem struct {
	Name                string
	Notes               []string
	RFunc               ResourceFunc
	CoreRFunc           CoreResourceFunc
	ReferenceAttributes []string
	CustomRefIDFunc     ReferenceIDFunc
	DefaultRefIDFunc    ReferenceIDFunc
	NoPrice             bool
}
