#!/usr/bin/env node

import inquirer from "inquirer";
import { info } from "./utils/logger";
import { getClient } from "./utils/ecs-client-provider";
import { Action } from "./functions/action";
import autocomplete from 'inquirer-autocomplete-prompt';
import fuzzy from 'fuzzy';
import { Execute } from "./functions/execute";
import { DescribeClustersCommand, DescribeTaskDefinitionCommand, DescribeTasksCommand, ListClustersCommand, ListTasksCommand } from "@aws-sdk/client-ecs";
import { inspect } from "util";
import { getTask } from "./prompts/get-task";
// TODO: Profile select
// TODO: Properly handle regions?

inquirer.registerPrompt('autocomplete', autocomplete);

async function selectAction() {
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

declare type ClusterEntry = { name: string, value: string; };
async function getCluster(): Promise<string> {
  info("Fetching cluster list...");
  const cmd = new ListClustersCommand({ maxResults: 100 });
  const clusters = await getClient().send(cmd);
  const clusterList: ClusterEntry[] = clusters.clusterArns?.map((arn) => ({ name: arn.split('/')[1], value: arn })) || [];
  const prompt = await inquirer
    .prompt([
      {
        type: 'autocomplete',
        name: 'instance',
        message: 'Select a cluster (type to filter)',
        source: (_: any, input: string): Promise<ClusterEntry[]> => {
          input = input || "";
          return new Promise((res): void => {
            const results = fuzzy.filter<ClusterEntry>(input, clusterList, { extract: (e): string => e.name });
            return res(results.map((e): ClusterEntry => e.original));
          });
        }
      }]);
  return prompt.instance;
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
