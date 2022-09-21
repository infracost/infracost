package schema

import (
	"time"
)

type ActualCosts struct {
	ResourceID     string
	StartTimestamp time.Time
	EndTimestamp   time.Time
	CostComponents []*CostComponent
}
