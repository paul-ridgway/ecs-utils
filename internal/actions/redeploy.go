package actions

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/paul-ridgway/ecs-utils/internal/prompts"
	"github.com/paul-ridgway/ecs-utils/internal/ui"
)

// RedeployService updates a service to a chosen task definition revision and
// forces a new deployment.
type RedeployService struct{}

// Name implements Action.
func (RedeployService) Name() string { return "Re-deploy Service" }

// Run implements Action.
func (RedeployService) Run(c *Context) error {
	cluster, err := c.ClusterArn()
	if err != nil {
		return err
	}
	service, err := c.Service()
	if err != nil {
		return err
	}
	taskDefinition, err := c.TaskDefinition(prompts.FamilyFromTaskDefinitionArn(aws.ToString(service.TaskDefinition)))
	if err != nil {
		return err
	}
	_, err = c.Session.ECS.UpdateService(c.Ctx, &ecs.UpdateServiceInput{
		Cluster:            aws.String(cluster),
		Service:            service.ServiceArn,
		TaskDefinition:     aws.String(taskDefinition.Arn),
		ForceNewDeployment: true,
	})
	if err != nil {
		return err
	}
	ui.Info("Service update command sent successfully")
	return nil
}
