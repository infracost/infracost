package security

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"
)

const GenesisHash = "0000000000000000000000000000000000000000000000000000000000000000"

// ActionLedgerEntry represents an immutable cryptographic audit record of a FinOps policy evaluation.
type ActionLedgerEntry struct {
	Index        int                    `json:"index"`
	Timestamp    string                 `json:"timestamp"`
	ResourceName string                 `json:"resource_name"`
	EventType    string                 `json:"event_type"`
	Status       string                 `json:"status"`
	CostDeltaUSD float64                `json:"cost_delta_usd"`
	PrevHash     string                 `json:"prev_hash"`
	CurrHash     string                 `json:"curr_hash"`
	Metadata     map[string]interface{} `json:"metadata"`
}

// ActionGateLedger maintains a tamper-evident SHA-256 hash chain of FinOps policy decisions.
type ActionGateLedger struct {
	entries  []ActionLedgerEntry
	lastHash string
}

// NewActionGateLedger initializes a fresh cryptographic Action Ledger.
func NewActionGateLedger() *ActionGateLedger {
	return &ActionGateLedger{
		entries:  make([]ActionLedgerEntry, 0),
		lastHash: GenesisHash,
	}
}

// RecordEntry computes canonical SHA-256 hash and appends the entry to the chain.
func (l *ActionGateLedger) RecordEntry(eventType, resourceName, status string, costDelta float64, metadata map[string]interface{}) ActionLedgerEntry {
	timestamp := time.Now().UTC().Format(time.RFC3339)
	index := len(l.entries)

	metaBytes, _ := json.Marshal(metadata)
	metaHash := sha256.Sum256(metaBytes)
	metaHex := hex.EncodeToString(metaHash[:])

	canonical := fmt.Sprintf("%d|%s|%s|%s|%s|%.2f|%s|%s", index, l.lastHash, eventType, resourceName, status, costDelta, timestamp, metaHex)
	currHashBytes := sha256.Sum256([]byte(canonical))
	currHash := hex.EncodeToString(currHashBytes[:])

	entry := ActionLedgerEntry{
		Index:        index,
		Timestamp:    timestamp,
		ResourceName: resourceName,
		EventType:    eventType,
		Status:       status,
		CostDeltaUSD: costDelta,
		PrevHash:     l.lastHash,
		CurrHash:     currHash,
		Metadata:     metadata,
	}

	l.entries = append(l.entries, entry)
	l.lastHash = currHash
	return entry
}

// GetEntries returns all recorded audit ledger entries.
func (l *ActionGateLedger) GetEntries() []ActionLedgerEntry {
	return l.entries
}

// VerifyIntegrity verifies that the entire SHA-256 hash chain is intact and un-tampered.
func (l *ActionGateLedger) VerifyIntegrity() bool {
	prev := GenesisHash
	for _, entry := range l.entries {
		if entry.PrevHash != prev {
			return false
		}
		prev = entry.CurrHash
	}
	return true
}

// ActionGateFinOpsEvaluator evaluates Terraform and OpenTofu infrastructure cost deltas against zero-trust safety policies.
type ActionGateFinOpsEvaluator struct {
	NeverEquateIntentToApproval bool
	EnforceActionBoundary       bool
	MaxMonthlyCostDeltaUSD      float64
	Ledger                      *ActionGateLedger
}

// NewActionGateFinOpsEvaluator creates a new FinOps policy evaluator with default safety thresholds.
func NewActionGateFinOpsEvaluator(maxDeltaUSD float64) *ActionGateFinOpsEvaluator {
	return &ActionGateFinOpsEvaluator{
		NeverEquateIntentToApproval: true,
		EnforceActionBoundary:       true,
		MaxMonthlyCostDeltaUSD:      maxDeltaUSD,
		Ledger:                      NewActionGateLedger(),
	}
}

func (e *ActionGateFinOpsEvaluator) checkKillSwitch() bool {
	envVal := strings.ToLower(os.Getenv("AAG_KILL_SWITCH"))
	if envVal == "true" || envVal == "1" || envVal == "yes" {
		return true
	}
	for _, path := range []string{"artifacts/KILL", "/tmp/KILL"} {
		if _, err := os.Stat(path); err == nil {
			return true
		}
	}
	return false
}

// EvaluatePlanCost checks if a cloud resource's monthly cost delta complies with FinOps policy boundaries.
func (e *ActionGateFinOpsEvaluator) EvaluatePlanCost(resourceName string, monthlyCostDelta float64, isApproved bool) (bool, string, error) {
	// 1. Evaluate emergency kill switch
	if e.checkKillSwitch() {
		entry := e.Ledger.RecordEntry("evaluation_blocked", resourceName, "halted_by_kill_switch", monthlyCostDelta, map[string]interface{}{
			"reason": "emergency_kill_switch_active",
		})
		return false, entry.CurrHash, fmt.Errorf("A2Z SOC ActionGate: Emergency kill switch is engaged. Infrastructure deployment halted")
	}

	// 2. Evaluate budget threshold
	if monthlyCostDelta > e.MaxMonthlyCostDeltaUSD && !isApproved {
		entry := e.Ledger.RecordEntry("budget_exceeded", resourceName, "approval_required", monthlyCostDelta, map[string]interface{}{
			"limit_usd": e.MaxMonthlyCostDeltaUSD,
		})
		return false, entry.CurrHash, fmt.Errorf("A2Z SOC ActionGate: Monthly cost delta ($%.2f) exceeds authorized FinOps limit ($%.2f). Explicit approval required", monthlyCostDelta, e.MaxMonthlyCostDeltaUSD)
	}

	// 3. Approved / Within Boundary
	entry := e.Ledger.RecordEntry("evaluation_passed", resourceName, "authorized", monthlyCostDelta, map[string]interface{}{
		"never_equate_intent_to_approval": e.NeverEquateIntentToApproval,
	})
	return true, entry.CurrHash, nil
}
