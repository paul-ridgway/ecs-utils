// Package prompts fetches ECS resources and asks the user to pick one.
package prompts

import "strings"

// resourceID returns the part of an ARN after the first "/", e.g. the cluster
// name from a cluster ARN or "family:revision" from a task definition ARN.
func resourceID(arn string) string {
	_, id, ok := strings.Cut(arn, "/")
	if !ok {
		return arn
	}
	return id
}

// FamilyFromTaskDefinitionArn extracts the family name from a task definition ARN.
func FamilyFromTaskDefinitionArn(arn string) string {
	if arn == "" {
		return ""
	}
	family, _, _ := strings.Cut(resourceID(arn), ":")
	return family
}

func chunk[T any](items []T, size int) [][]T {
	var out [][]T
	for len(items) > 0 {
		n := min(size, len(items))
		out = append(out, items[:n])
		items = items[n:]
	}
	return out
}
