package datami

import "testing"

func TestMediaZeroValue(t *testing.T) {
	m := Media{}
	if m.Titles != nil {
		t.Fatal("Titles should be nil on zero value")
	}
	if m.Raw != nil {
		t.Fatal("Raw should be nil on zero value")
	}
}
