package prompts

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/paul-ridgway/ecs-utils/internal/ui"
)

// TaskDefinitionSummary is one revision of a task definition family.
type TaskDefinitionSummary struct {
	Name     string
	Arn      string
	Revision int
	Latest   bool
}

// TaskDefinition lists the active task definitions (optionally limited to a
// family prefix), newest first, and asks the user to pick one.
func TaskDefinition(ctx context.Context, client *ecs.Client, familyPrefix string) (TaskDefinitionSummary, error) {
	ui.Info("Fetching task definition list...")
	input := &ecs.ListTaskDefinitionsInput{}
	if familyPrefix != "" {
		input.FamilyPrefix = aws.String(familyPrefix)
	}
	var defs []TaskDefinitionSummary
	paginator := ecs.NewListTaskDefinitionsPaginator(client, input)
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return TaskDefinitionSummary{}, err
		}
		for _, arn := range page.TaskDefinitionArns {
			defs = append(defs, extractDetails(arn))
		}
	}
	if len(defs) == 0 {
		return TaskDefinitionSummary{}, errors.New("no task definitions found")
	}
	sort.SliceStable(defs, func(i, j int) bool { return defs[i].Revision > defs[j].Revision })
	markLatest(defs)

	options := make([]ui.Option[TaskDefinitionSummary], 0, len(defs))
	for _, def := range defs {
		label := fmt.Sprintf("%s (%d)", def.Name, def.Revision)
		if def.Latest {
			label += " [LATEST]"
		}
		options = append(options, ui.Option[TaskDefinitionSummary]{Label: label, Value: def})
	}
	return ui.Select("Select a task definition (type to filter)", options)
}

// markLatest flags the highest revision of each family.
func markLatest(defs []TaskDefinitionSummary) {
	latest := map[string]int{}
	for i, def := range defs {
		if cur, ok := latest[def.Name]; !ok || def.Revision > defs[cur].Revision {
			latest[def.Name] = i
		}
	}
	for _, i := range latest {
		defs[i].Latest = true
	}
}

func extractDetails(arn string) TaskDefinitionSummary {
	name, rev, _ := strings.Cut(resourceID(arn), ":")
	revision, _ := strconv.Atoi(rev)
	return TaskDefinitionSummary{Arn: arn, Name: name, Revision: revision}
}
