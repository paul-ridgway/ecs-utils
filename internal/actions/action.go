// Package actions implements the operations the tool can perform.
package actions

import (
	"context"
	"sort"

	"github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/paul-ridgway/ecs-utils/internal/awsx"
	"github.com/paul-ridgway/ecs-utils/internal/prompts"
)

// Action is something the user can choose to do.
type Action interface {
	Name() string
	Run(c *Context) error
}

// All returns every available action, sorted by name.
func All() []Action {
	actions := []Action{Execute{}, RedeployService{}, ScaleService{}}
	sort.Slice(actions, func(i, j int) bool { return actions[i].Name() < actions[j].Name() })
	return actions
}

// Context carries the AWS session and remembers the cluster once chosen.
type Context struct {
	Ctx        context.Context
	Session    *awsx.Session
	clusterArn string
}

// NewContext creates a Context for the given session.
func NewContext(ctx context.Context, session *awsx.Session) *Context {
	return &Context{Ctx: ctx, Session: session}
}

// ClusterArn asks for a cluster the first time and caches the answer.
func (c *Context) ClusterArn() (string, error) {
	if c.clusterArn == "" {
		arn, err := prompts.Cluster(c.Ctx, c.Session.ECS)
		if err != nil {
			return "", err
		}
		c.clusterArn = arn
	}
	return c.clusterArn, nil
}

// Service asks for a service in the chosen cluster.
func (c *Context) Service() (types.Service, error) {
	cluster, err := c.ClusterArn()
	if err != nil {
		return types.Service{}, err
	}
	return prompts.Service(c.Ctx, c.Session.ECS, cluster)
}

// Task asks for a task and container in the chosen cluster.
func (c *Context) Task() (prompts.TaskContainer, error) {
	cluster, err := c.ClusterArn()
	if err != nil {
		return prompts.TaskContainer{}, err
	}
	return prompts.Task(c.Ctx, c.Session.ECS, cluster)
}

// TaskDefinition asks for a task definition, optionally within a family.
func (c *Context) TaskDefinition(familyPrefix string) (prompts.TaskDefinitionSummary, error) {
	return prompts.TaskDefinition(c.Ctx, c.Session.ECS, familyPrefix)
}
