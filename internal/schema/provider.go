package schema

type Provider interface {
	Type() string
	DisplayType() string
	AddMetadata(*ProjectMetadata)
	LoadResources(*Project, map[string]*UsageData) error
}
