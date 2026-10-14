package prompts

import "testing"

func TestResourceID(t *testing.T) {
	cases := map[string]string{
		"arn:aws:ecs:eu-west-2:123:cluster/my-cluster":        "my-cluster",
		"arn:aws:ecs:eu-west-2:123:task-definition/web:42":    "web:42",
		"arn:aws:ecs:eu-west-2:123:task/my-cluster/abc123def": "my-cluster/abc123def",
		"no-slash": "no-slash",
	}
	for in, want := range cases {
		if got := resourceID(in); got != want {
			t.Errorf("resourceID(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestFamilyFromTaskDefinitionArn(t *testing.T) {
	if got := FamilyFromTaskDefinitionArn("arn:aws:ecs:eu-west-2:123:task-definition/web:42"); got != "web" {
		t.Errorf("got %q, want web", got)
	}
	if got := FamilyFromTaskDefinitionArn(""); got != "" {
		t.Errorf("got %q, want empty", got)
	}
}

func TestMarkLatest(t *testing.T) {
	defs := []TaskDefinitionSummary{
		extractDetails("arn:aws:ecs:eu-west-2:123:task-definition/web:3"),
		extractDetails("arn:aws:ecs:eu-west-2:123:task-definition/web:2"),
		extractDetails("arn:aws:ecs:eu-west-2:123:task-definition/worker:9"),
		extractDetails("arn:aws:ecs:eu-west-2:123:task-definition/web:1"),
	}
	markLatest(defs)
	want := []bool{true, false, true, false}
	for i, def := range defs {
		if def.Latest != want[i] {
			t.Errorf("defs[%d] (%s:%d) latest = %v, want %v", i, def.Name, def.Revision, def.Latest, want[i])
		}
	}
}

func TestChunk(t *testing.T) {
	got := chunk([]int{1, 2, 3, 4, 5}, 2)
	if len(got) != 3 || len(got[0]) != 2 || len(got[2]) != 1 {
		t.Errorf("unexpected chunks: %v", got)
	}
	if chunk([]int{}, 2) != nil {
		t.Error("expected nil for empty input")
	}
}
