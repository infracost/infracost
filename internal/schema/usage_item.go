package schema

type UsageVariableType int

const (
	Int64 UsageVariableType = iota
	String
	Float64
	StringArray
	SubResourceUsage
	KeyValueMap
)

type UsageItem struct {
	Key          string
	DefaultValue any
	Value        any
	ValueType    UsageVariableType
	Description  string
}
