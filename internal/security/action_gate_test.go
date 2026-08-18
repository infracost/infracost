package security

import (
	"testing"
)

func TestActionGateFinOpsEvaluator_EvaluatePlanCost(t *testing.T) {
	evaluator := NewActionGateFinOpsEvaluator(500.0) // $500 monthly limit

	// 1. Within budget
	allowed, hash1, err := evaluator.EvaluatePlanCost("aws_instance.web", 150.0, false)
	if err != nil || !allowed || hash1 == "" {
		t.Fatalf("Expected allowed plan cost within budget, got err: %v", err)
	}

	// 2. Exceeds budget without approval
	allowed, _, err = evaluator.EvaluatePlanCost("aws_eks_cluster.prod", 1200.0, false)
	if err == nil || allowed {
		t.Fatalf("Expected budget exceeded error, got allowed")
	}

	// 3. Exceeds budget with explicit approval
	allowed, hash3, err := evaluator.EvaluatePlanCost("aws_eks_cluster.prod", 1200.0, true)
	if err != nil || !allowed || hash3 == "" {
		t.Fatalf("Expected approved plan cost to pass, got err: %v", err)
	}

	// 4. Verify cryptographic hash chain integrity
	entries := evaluator.Ledger.GetEntries()
	if len(entries) != 3 {
		t.Fatalf("Expected 3 ledger entries, got %d", len(entries))
	}
	if !evaluator.Ledger.VerifyIntegrity() {
		t.Fatalf("Expected ledger integrity verification to pass")
	}
}
