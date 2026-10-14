package actions

import (
	"errors"
	"os"
	"os/exec"
	"strings"

	"github.com/paul-ridgway/ecs-utils/internal/ui"
)

// Execute opens an interactive command in a running container using
// "aws ecs execute-command".
type Execute struct{}

// Name implements Action.
func (Execute) Name() string { return "Execute Command" }

// Run implements Action.
func (Execute) Run(c *Context) error {
	cluster, err := c.ClusterArn()
	if err != nil {
		return err
	}
	task, err := c.Task()
	if err != nil {
		return err
	}
	command, err := ui.Input("Command to run", "/bin/sh")
	if err != nil {
		return err
	}

	awsPath, err := exec.LookPath("aws")
	if err != nil {
		return errors.New("the AWS CLI is required for Execute Command but was not found on PATH (run ecs-utils --install-deps)")
	}
	if _, err := exec.LookPath("session-manager-plugin"); err != nil {
		ui.Warn("session-manager-plugin was not found on PATH; the session will probably fail (run ecs-utils --install-deps)")
	}

	args := []string{
		"ecs", "execute-command",
		"--cluster", cluster,
		"--task", task.TaskArn,
		"--container", task.Container,
		"--command", command,
		"--interactive",
		"--region", c.Session.Region,
	}
	if c.Session.Profile != "" {
		args = append(args, "--profile", c.Session.Profile)
	}

	segments := []string{
		"AWS SSM SESSION",
		"cluster: " + lastSegment(cluster),
		"task: " + lastSegment(task.TaskArn),
		"container: " + task.Container,
		"region: " + c.Session.Region,
	}
	if c.Session.Profile != "" {
		segments = append(segments, "profile: "+c.Session.Profile)
	}

	ui.Info("Note: Press ctrl-c twice within a second to exit, once to send to the container")
	cmd := exec.Command(awsPath, args...)
	code, err := runWithStatusBar(cmd, terminalIO{in: os.Stdin, out: os.Stdout}, segments)
	if errors.Is(err, errNoStatusBar) {
		code, err = runPlain(exec.Command(awsPath, args...))
	}
	if err != nil {
		ui.Error("Shell error: %v", err)
		return err
	}
	ui.Info("Process terminated with code: %d", code)
	return nil
}

// lastSegment returns the part of an ARN after its final "/".
func lastSegment(arn string) string {
	if i := strings.LastIndex(arn, "/"); i >= 0 {
		return arn[i+1:]
	}
	return arn
}
