# ecs-utils

A little utility to quickly manage ECS clusters, services and tasks from the command line.
Everything is interactive: pick an action, pick a cluster, type to filter, done.

## What it does

- **Execute Command** – open an interactive shell (or any command) in a running container, via `aws ecs execute-command`. While the session is open a status bar is pinned to the bottom of the terminal showing that you are in an AWS SSM session, plus the cluster, task, container, region and profile. Press ctrl-c once to send it to the container, twice within a second to end the session. (The status bar needs a real terminal and is not shown on Windows, where the session runs directly in the console.)
- **Re-deploy Service** – pick a service, pick a task definition revision from its family (the newest is marked `[LATEST]`) and force a new deployment.
- **Scale Service** – pick a service, see its current pending/running/desired counts and set a new desired count.

On start-up it also:

- checks for the **AWS CLI** and the **Session Manager plugin** and offers to install whichever is missing (see below);
- asks which **AWS profile** to use when neither `--profile` nor `AWS_PROFILE` is set, listing the profiles from `~/.aws/config` and `~/.aws/credentials`;
- offers to run `aws sso login` when the chosen profile uses IAM Identity Center and its session has expired.

## Install

Download a binary for your platform from the [releases page](https://github.com/paul-ridgway/ecs-utils/releases) and put it on your `PATH`, or build from source:

```sh
go install github.com/paul-ridgway/ecs-utils@latest
```

## Usage

```sh
ecs-utils                      # interactive, prompts for anything it needs
ecs-utils --profile prod       # use a specific profile
ecs-utils --region eu-west-1   # override the region
ecs-utils --install-deps       # only check/install the AWS CLI and plugin, then exit
ecs-utils --install-deps -y    # same, without asking first
ecs-utils --skip-deps          # do not check for dependencies at all
ecs-utils --version
```

The region is taken from `--region`, then `AWS_REGION`/`AWS_DEFAULT_REGION`, then the profile; if none is set it falls back to `eu-west-2`.

## Dependency installation

`Execute Command` needs the AWS CLI (v2) and the Session Manager plugin. When either is missing ecs-utils can install it for you without root:

| Platform | AWS CLI | Session Manager plugin |
| --- | --- | --- |
| Linux | Official bundle, installed to `~/.local/aws-cli` with a launcher in `~/.local/bin` | Official package unpacked to `~/.local/sessionmanagerplugin`, linked from `~/.local/bin` |
| macOS | Homebrew if available, otherwise the official `.pkg` installed into your home directory | Homebrew if available, otherwise the official bundle unpacked to `~/.local/sessionmanagerplugin` |
| Windows | winget if available, otherwise the official MSI | winget if available, otherwise the official installer |

If `~/.local/bin` is not on your `PATH` you will be told how to add it; ecs-utils itself will still find the tools there.

## Development

```sh
go build ./...
go test ./...
go run . --skip-deps
```

Releases are built with [GoReleaser](https://goreleaser.com) when a `v*` tag is pushed.
