package ui

import "testing"

func TestFuzzyMatch(t *testing.T) {
	cases := []struct {
		pattern, s string
		want       bool
	}{
		{"", "anything", true},
		{"prd", "production-cluster", true},
		{"PRD", "production-cluster", true},
		{"xyz", "production-cluster", false},
		{"redeploy", "Re-deploy Service", true},
		{"sclsrv", "Scale Service", true},
		{"service scale", "Scale Service", false},
	}
	for _, c := range cases {
		if got := FuzzyMatch(c.pattern, c.s); got != c.want {
			t.Errorf("FuzzyMatch(%q, %q) = %v, want %v", c.pattern, c.s, got, c.want)
		}
	}
}
