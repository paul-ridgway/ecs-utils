package prompts

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/paul-ridgway/ecs-utils/internal/ui"
)

// Cluster lists the clusters in the region and asks the user to pick one,
// returning its ARN.
func Cluster(ctx context.Context, client *ecs.Client) (string, error) {
	ui.Info("Fetching cluster list...")
	var options []ui.Option[string]
	paginator := ecs.NewListClustersPaginator(client, &ecs.ListClustersInput{MaxResults: aws.Int32(100)})
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return "", err
		}
		for _, arn := range page.ClusterArns {
			options = append(options, ui.Option[string]{Label: resourceID(arn), Value: arn})
		}
	}
	if len(options) == 0 {
		return "", errors.New("no ECS clusters found in this region")
	}
	return ui.Select("Select a cluster (type to filter)", options)
}
