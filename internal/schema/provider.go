package schema

type Provider interface {
	Type() string
	DisplayType() string
	AddMetadata(*ProjectMetadata)
	LoadResources(map[string]*UsageData) ([]*Project, error)
}
