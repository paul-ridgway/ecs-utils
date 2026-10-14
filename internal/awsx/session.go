package awsx

import (
	"context"
	"fmt"
	"os"
	"os/exec"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/paul-ridgway/ecs-utils/internal/ui"
)

// DefaultRegion is used when neither the flags, environment nor profile set one.
const DefaultRegion = "eu-west-2"

// Session bundles the resolved AWS settings and the ECS client.
type Session struct {
	ECS     *ecs.Client
	Profile string
	Region  string
}

// NewSession loads AWS configuration for the profile and region, verifies that
// credentials can be obtained (offering an SSO login where appropriate) and
// returns a ready-to-use ECS client.
func NewSession(ctx context.Context, profile, region string) (*Session, error) {
	if profile != "" {
		// Child processes (the AWS CLI) should see the same profile.
		_ = os.Setenv("AWS_PROFILE", profile)
	}

	cfg, err := loadConfig(ctx, profile, region)
	if err != nil {
		return nil, err
	}

	if _, err := cfg.Credentials.Retrieve(ctx); err != nil {
		if p, ok := LookupProfile(profile); ok && p.IsSSO() {
			ui.Warn("Could not load credentials for profile %q: %v", profile, err)
			cfg, err = ssoLogin(ctx, profile, region)
		}
		if err != nil {
			return nil, fmt.Errorf("unable to load AWS credentials: %w", err)
		}
	}

	if profile != "" {
		ui.Info("Using profile %s in region %s", profile, cfg.Region)
	} else {
		ui.Info("Using region %s", cfg.Region)
	}

	return &Session{ECS: ecs.NewFromConfig(cfg), Profile: profile, Region: cfg.Region}, nil
}

func loadConfig(ctx context.Context, profile, region string) (aws.Config, error) {
	var opts []func(*config.LoadOptions) error
	if profile != "" {
		opts = append(opts, config.WithSharedConfigProfile(profile))
	}
	if region != "" {
		opts = append(opts, config.WithRegion(region))
	}
	cfg, err := config.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		return aws.Config{}, fmt.Errorf("loading AWS config: %w", err)
	}
	if cfg.Region == "" {
		ui.Warn("No region configured; defaulting to %s (use --region or set AWS_REGION to change)", DefaultRegion)
		cfg.Region = DefaultRegion
	}
	return cfg, nil
}

// ssoLogin runs "aws sso login" for the profile and reloads the configuration.
func ssoLogin(ctx context.Context, profile, region string) (aws.Config, error) {
	ok, err := ui.Confirm(fmt.Sprintf("Run 'aws sso login --profile %s' now?", profile), true)
	if err != nil {
		return aws.Config{}, err
	}
	if !ok {
		return aws.Config{}, fmt.Errorf("profile %q needs an SSO login", profile)
	}
	awsPath, err := exec.LookPath("aws")
	if err != nil {
		return aws.Config{}, fmt.Errorf("the AWS CLI is needed to log in with SSO but was not found on PATH")
	}
	cmd := exec.CommandContext(ctx, awsPath, "sso", "login", "--profile", profile)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	if err := cmd.Run(); err != nil {
		return aws.Config{}, fmt.Errorf("aws sso login failed: %w", err)
	}

	cfg, err := loadConfig(ctx, profile, region)
	if err != nil {
		return aws.Config{}, err
	}
	if _, err := cfg.Credentials.Retrieve(ctx); err != nil {
		return aws.Config{}, err
	}
	return cfg, nil
}
