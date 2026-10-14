package actions

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/paul-ridgway/ecs-utils/internal/ui"
)

// ScaleService changes the desired count of a service.
type ScaleService struct{}

// Name implements Action.
func (ScaleService) Name() string { return "Scale Service" }

// Run implements Action.
func (ScaleService) Run(c *Context) error {
	cluster, err := c.ClusterArn()
	if err != nil {
		return err
	}
	service, err := c.Service()
	if err != nil {
		return err
	}
	ui.Info("Currently, pending: %d, running: %d of desired: %d", service.PendingCount, service.RunningCount, service.DesiredCount)
	count, err := ui.InputInt("Desired count", int(service.DesiredCount))
	if err != nil {
		return err
	}
	_, err = c.Session.ECS.UpdateService(c.Ctx, &ecs.UpdateServiceInput{
		Cluster:            aws.String(cluster),
		Service:            service.ServiceArn,
		DesiredCount:       aws.Int32(int32(count)),
		ForceNewDeployment: true,
	})
	if err != nil {
		return err
	}
	ui.Info("Service update command sent successfully")
	return nil
}
