package schema

// ReferenceIDFunc is used to let references be built using non-standard IDs (anything other than d.Get("id").string)
type ReferenceIDFunc func(d *ResourceData) []string

// CloudResourceIDFunc is used to calculate the cloud resource ids (AWS ARN, Google HREF, etc...) associated with the resource
type CloudResourceIDFunc func(d *ResourceData) []string

type RegistryItem struct {
	Name                string
	Notes               []string
	RFunc               ResourceFunc
	CoreRFunc           CoreResourceFunc
	ReferenceAttributes []string
	CustomRefIDFunc     ReferenceIDFunc
	DefaultRefIDFunc    ReferenceIDFunc
	CloudResourceIDFunc CloudResourceIDFunc
	NoPrice             bool
}
