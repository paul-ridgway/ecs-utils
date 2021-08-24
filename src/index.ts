#!/usr/bin/env node

import inquirer from "inquirer";
import { Action } from "./functions/action";
import autocomplete from 'inquirer-autocomplete-prompt';
import { Execute } from "./functions/execute";
import { getTask } from "./prompts/get-task";
import { getCluster } from "./prompts/get-cluster";
// TODO: Profile select
// TODO: Properly handle regions?

inquirer.registerPrompt('autocomplete', autocomplete);

async function selectAction(): Promise<any> {
  const prompt = await inquirer
    .prompt([{
      type: 'list',
      name: 'action',
      loop: false,
      message: 'What do you want to do?',
      choices: [
        { name: 'Execute Command', value: Execute },
      ],
    },
    ]);
  return prompt.action;
}
// TODO: Move these to function files

function getService(): string {
  return "";
}

async function run(): Promise<void> {
  const Act = await selectAction();
  const act: Action = new Act(getCluster, getService, getTask);
  await act.run();
}

run()
  .then((): void => process.exit(0))
  .catch((err): void => { throw (err); });
