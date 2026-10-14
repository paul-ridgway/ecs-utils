package prompts

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/paul-ridgway/ecs-utils/internal/ui"
)

// TaskContainer identifies one container inside a running task.
type TaskContainer struct {
	TaskArn   string
	Container string
}

// Task lists the tasks in a cluster, expands them to their containers and asks
// the user to pick one.
func Task(ctx context.Context, client *ecs.Client, clusterArn string) (TaskContainer, error) {
	ui.Info("Fetching task list...")
	var arns []string
	paginator := ecs.NewListTasksPaginator(client, &ecs.ListTasksInput{Cluster: aws.String(clusterArn)})
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return TaskContainer{}, err
		}
		arns = append(arns, page.TaskArns...)
	}
	if len(arns) == 0 {
		return TaskContainer{}, errors.New("no running tasks found in this cluster")
	}

	var options []ui.Option[TaskContainer]
	for _, batch := range chunk(arns, 100) { // DescribeTasks accepts at most 100
		out, err := client.DescribeTasks(ctx, &ecs.DescribeTasksInput{Cluster: aws.String(clusterArn), Tasks: batch})
		if err != nil {
			return TaskContainer{}, err
		}
		for _, task := range out.Tasks {
			for _, container := range task.Containers {
				label := fmt.Sprintf("Task Definition: %s, Container: %s (%s)",
					resourceID(aws.ToString(task.TaskDefinitionArn)), aws.ToString(container.Name), aws.ToString(container.LastStatus))
				options = append(options, ui.Option[TaskContainer]{
					Label: label,
					Value: TaskContainer{TaskArn: aws.ToString(task.TaskArn), Container: aws.ToString(container.Name)},
				})
			}
		}
	}
	return ui.Select("Select a task and container (type to filter)", options)
}
