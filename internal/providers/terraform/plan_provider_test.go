package terraform

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFindTerraformDir(t *testing.T) {
	tmp, err := os.MkdirTemp("", "infracost_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmp)

	parent := filepath.Join(tmp, "parent")
	sub := filepath.Join(parent, "sub")
	if err := os.MkdirAll(sub, 0755); err != nil {
		t.Fatal(err)
	}

	// create a file that IsTerraformDir should detect (e.g., main.tf)
	tfFile := filepath.Join(parent, "main.tf")
	if err := os.WriteFile(tfFile, []byte("resource \"aws_s3_bucket\" \"b\" {}"), 0644); err != nil {
		t.Fatal(err)
	}

	found, ok := findTerraformDir(sub)
	if !ok {
		t.Fatalf("expected to find terraform dir but did not")
	}
	if found != parent {
		t.Fatalf("expected found dir %s, got %s", parent, found)
	}

	// directory with no terraform files
	other := filepath.Join(tmp, "other")
	if err := os.MkdirAll(other, 0755); err != nil {
		t.Fatal(err)
	}
	_, ok = findTerraformDir(other)
	if ok {
		t.Fatalf("expected not to find terraform dir")
	}
}

func TestFindTerraformDir_DirectDir(t *testing.T) {
	tmp, err := os.MkdirTemp("", "infracost_test_dir")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmp)

	tfFile := filepath.Join(tmp, "main.tf")
	if err := os.WriteFile(tfFile, []byte("resource \"aws_s3_bucket\" \"b\" {}"), 0644); err != nil {
		t.Fatal(err)
	}

	found, ok := findTerraformDir(tmp)
	if !ok {
		t.Fatalf("expected to find terraform dir in same directory")
	}
	if found != tmp {
		t.Fatalf("expected found dir %s, got %s", tmp, found)
	}
}
