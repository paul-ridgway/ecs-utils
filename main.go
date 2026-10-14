// Command ecs-utils is a small interactive helper for managing ECS clusters,
// services and tasks from the command line.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/paul-ridgway/ecs-utils/internal/actions"
	"github.com/paul-ridgway/ecs-utils/internal/awsx"
	"github.com/paul-ridgway/ecs-utils/internal/deps"
	"github.com/paul-ridgway/ecs-utils/internal/ui"
)

// Set by the release build via -ldflags.
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

type options struct {
	profile     string
	region      string
	skipDeps    bool
	installDeps bool
	assumeYes   bool
}

func main() {
	var opts options
	var showVersion bool
	flag.StringVar(&opts.profile, "profile", "", "AWS profile to use (defaults to $AWS_PROFILE; prompts if neither is set)")
	flag.StringVar(&opts.region, "region", "", "AWS region (defaults to $AWS_REGION or the profile's region, then "+awsx.DefaultRegion+")")
	flag.BoolVar(&opts.skipDeps, "skip-deps", false, "do not check for the AWS CLI and Session Manager plugin")
	flag.BoolVar(&opts.installDeps, "install-deps", false, "check for and install missing dependencies, then exit")
	flag.BoolVar(&opts.assumeYes, "y", false, "install missing dependencies without asking")
	flag.BoolVar(&showVersion, "version", false, "print the version and exit")
	flag.Usage = usage
	flag.Parse()

	if showVersion {
		fmt.Printf("ecs-utils %s (commit %s, built %s)\n", version, commit, date)
		return
	}

	if err := run(context.Background(), opts); err != nil {
		if ui.IsInterrupt(err) {
			fmt.Println()
			os.Exit(130)
		}
		ui.Error("%v", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, opts options) error {
	if !opts.skipDeps || opts.installDeps {
		if err := deps.Ensure(ctx, opts.assumeYes); err != nil {
			return err
		}
	}
	if opts.installDeps {
		return nil
	}

	profile, err := awsx.ResolveProfile(opts.profile)
	if err != nil {
		return err
	}
	session, err := awsx.NewSession(ctx, profile, opts.region)
	if err != nil {
		return err
	}

	action, err := selectAction()
	if err != nil {
		return err
	}
	return action.Run(actions.NewContext(ctx, session))
}

func selectAction() (actions.Action, error) {
	var options []ui.Option[actions.Action]
	for _, action := range actions.All() {
		options = append(options, ui.Option[actions.Action]{Label: action.Name(), Value: action})
	}
	return ui.Select("What do you want to do?", options)
}

func usage() {
	fmt.Fprintf(flag.CommandLine.Output(), `ecs-utils %s - manage ECS clusters, services and tasks from the command line

Usage: ecs-utils [flags]

Flags:
`, version)
	flag.PrintDefaults()
}
