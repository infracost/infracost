package azurerm

import (
	"encoding/json"
	"log"

	"github.com/tidwall/gjson"
)

type ChangeType string
type PropertyChangeType string

const (
	Create      ChangeType = "Create"
	Delete                 = "Delete"
	Deploy                 = "Deploy"
	Ignore                 = "Ignore"
	Modify                 = "Modify"
	NoChange               = "NoChange"
	Unsupported            = "Unsupported"
)

const (
	PropCreate   PropertyChangeType = "Create"
	PropDelete                      = "Delete"
	PropArray                       = "Array"
	PropModify                      = "Modify"
	PropNoEffect                    = "NoEffect"
)

// Struct for serializing the JSON response of a whatif call
// Modeled after the schema of deployments/whatIf in the AzureRM REST API
// see: https://learn.microsoft.com/en-us/rest/api/resources/deployments/what-if-at-subscription-scope
type WhatIf struct {
	Status  string             `json:"status"`
	Error   ErrorResponse      `json:"error,omitempty"`
	Changes []ResourceSnapshot `json:"changes,omitempty"`
}

type WhatifProperties struct {
	CorrelationId string `json:"correlationId"`
}

type ResourceSnapshot struct {
	ResourceId        string     `json:"resourceId"`
	UnsupportedReason string     `json:"unsupportedReason,omitempty"`
	ChangeType        ChangeType `json:"changeType"`

	// Before/After include several fields that are always present (resourceId, type etc.)
	// A resource's 'properties' field differs greatly, so serialize as raw JSON
	BeforeRaw json.RawMessage `json:"before,omitempty"`
	AfterRaw  json.RawMessage `json:"after,omitempty"`
	// TODO: Should be of type WhatIfChange
	DeltaRaw json.RawMessage `json:"delta,omitempty"`

	// Parsed backing fields using gjson

	before gjson.Result
	after  gjson.Result
}

func (w *ResourceSnapshot) After() gjson.Result {
	if w.after.Get("id").Exists() {
		return w.after
	}

	marshal, err := w.AfterRaw.MarshalJSON()
	if err != nil {
		log.Fatalf("Failed marshalling After")
	}

	w.after = gjson.ParseBytes(marshal)
	return w.after
}

func (w *ResourceSnapshot) Before() gjson.Result {
	if w.before.Get("id").Exists() {
		return w.before
	}

	marshal, err := w.BeforeRaw.MarshalJSON()
	if err != nil {
		log.Fatalf("Failed marshalling Before")
	}

	w.before = gjson.ParseBytes(marshal)

	return w.before
}

type WhatIfPropertyChange struct {
	After  json.RawMessage `json:"after,omitempty"`
	Before json.RawMessage `json:"before,omitempty"`
	// TODO: this should be of type []WhatIfPropertyChange
	// go lang structs can't do self-referring, references work?
	Children []*WhatIfPropertyChange `json:"children,omitempty"`
	Path     string                  `json:"path,omitempty"`
}

type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Target  string `json:"target"`
}

type ErrorAdditionalInfo struct {
	Info string `json:"info"`
	Type string `json:"type"`
}
