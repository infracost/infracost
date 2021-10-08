package schema

type UsageVariableType int

const (
	Int64 UsageVariableType = iota
	String
	Float64
	StringArray
	Items
)

// type UsageDataValidatorFuncType = func(value interface{}) error

type UsageItem struct {
	Key          string
	DefaultValue interface{}
	Value        interface{}
	ValueType    UsageVariableType
	Description  string
	// These aren't used yet and I'm not entirely sure how they fit in, but they were part of the discussion about usage schema.
	// ValidatorFunc UsageDataValidatorFuncType
}
