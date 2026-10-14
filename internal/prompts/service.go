package prompts

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/paul-ridgway/ecs-utils/internal/ui"
)

// Service lists the services in a cluster and asks the user to pick one.
func Service(ctx context.Context, client *ecs.Client, clusterArn string) (types.Service, error) {
	ui.Info("Fetching service list...")
	var arns []string
	paginator := ecs.NewListServicesPaginator(client, &ecs.ListServicesInput{Cluster: aws.String(clusterArn)})
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return types.Service{}, err
		}
		arns = append(arns, page.ServiceArns...)
	}
	if len(arns) == 0 {
		return types.Service{}, errors.New("no services found in this cluster")
	}

	var options []ui.Option[types.Service]
	for _, batch := range chunk(arns, 10) { // DescribeServices accepts at most 10
		out, err := client.DescribeServices(ctx, &ecs.DescribeServicesInput{Cluster: aws.String(clusterArn), Services: batch})
		if err != nil {
			return types.Service{}, err
		}
		for _, svc := range out.Services {
			label := fmt.Sprintf("Name: %s, Task Definition: %s", aws.ToString(svc.ServiceName), resourceID(aws.ToString(svc.TaskDefinition)))
			options = append(options, ui.Option[types.Service]{Label: label, Value: svc})
		}
	}
	return ui.Select("Select a service (type to filter)", options)
}
