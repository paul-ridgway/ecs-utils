#!/usr/bin/env node

import inquirer from "inquirer";
import { Action } from "./functions/Action";
import autocomplete from 'inquirer-autocomplete-prompt';
import { Execute } from "./functions/Execute";
import { getTask } from "./prompts/getTask";
import { getCluster } from "./prompts/getCluster";
import { RedeployService } from "./functions/RedeployService";
import { getService } from "./prompts/getService";
import { ScaleService } from "./functions/ScaleService";
// TODO: Profile select
// TODO: Properly handle regions?

inquirer.registerPrompt('autocomplete', autocomplete);

type FnOption = {
  name: string;
  value: new (...args: any[]) => Action;
};

const fns: FnOption[] = [
  { name: 'Execute Command', value: Execute },
  { name: 'Re-deploy Service', value: RedeployService },
  { name: 'Scale Service', value: ScaleService },
].sort((a: FnOption, b: FnOption): number => a.name.localeCompare(b.name));

async function selectAction(): Promise<any> {
  const prompt = await inquirer
    .prompt([{
      type: 'autocomplete',
      name: 'action',
      message: 'What do you want to do?',
      source: async (answersSoFar: any, input: string): Promise<any[]> => {
        input = input || '';
        return fns.filter((fn): boolean => fn.name.toLowerCase().includes(input.toLowerCase()));
      },
    },
    ]);
  return prompt.action;
}

async function run(): Promise<void> {
  const Act = await selectAction();
  const act: Action = new Act(getCluster, getService, getTask);
  await act.run();
}

run()
  .then((): void => process.exit(0))
  .catch((err): void => { throw (err); });
