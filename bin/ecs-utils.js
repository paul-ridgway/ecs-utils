#!/usr/bin/env node
"use strict";

// This npm package is retired. ecs-utils is now a single Go binary with no
// Node.js dependency; this stub only tells you how to get it.

const bold = (s) => `\u001b[1m${s}\u001b[0m`;
const yellow = (s) => `\u001b[33m${s}\u001b[0m`;
const cyan = (s) => `\u001b[36m${s}\u001b[0m`;

const lines = [
  "",
  yellow(bold("ecs-utils has moved to a native Go binary and this npm package is no longer maintained.")),
  "",
  "The new version keeps the same workflow (Execute Command, Re-deploy Service, Scale Service),",
  "adds an AWS profile picker, installs the AWS CLI and Session Manager plugin for you, and",
  "shows a status bar while you are in a container session.",
  "",
  bold("Install it:"),
  "",
  `  Download a binary for your platform from`,
  `    ${cyan("https://github.com/paul-ridgway/ecs-utils/releases/latest")}`,
  `  and put it on your PATH, or build from source with Go:`,
  `    ${cyan("go install github.com/paul-ridgway/ecs-utils@latest")}`,
  "",
  bold("Then remove this package:"),
  "",
  `    ${cyan("npm uninstall -g ecs-utils")}`,
  "",
];

process.stderr.write(lines.join("\n") + "\n");
process.exit(1);
