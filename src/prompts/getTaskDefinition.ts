import { ListTaskDefinitionsCommand } from "@aws-sdk/client-ecs";
import { getClient } from "../utils/ecs-client-provider";
import { info } from "../utils/logger";
import inquirer from "inquirer";
import fuzzy from 'fuzzy';

export declare type TaskDefinitionSummary = {
  name: string,
  arn: string,
  revision: number;
  latest?: boolean;
};
export declare type TaskEntry = { name: string, value: TaskDefinitionSummary; };

export async function getTaskDefinition(familyPrefix: string | undefined): Promise<TaskDefinitionSummary> {
  info("Fetching task definition list...");
  const client = getClient();
  const taskDefinitionArns = [];
  let nextToken: string | undefined;
  do {
    const arns = await client.send(new ListTaskDefinitionsCommand({ familyPrefix, nextToken }));
    nextToken = arns.nextToken;
    if (arns.taskDefinitionArns) {
      taskDefinitionArns.push(...arns.taskDefinitionArns);
    }
  } while(nextToken)

  const taskDefinitions = markLatest((taskDefinitionArns || [])
    .map(extractDetails)
    .sort((a, b): number => b.revision - a.revision));

  const taskList: TaskEntry[] = taskDefinitions.map((cont): TaskEntry => {
    return {
      name: `${cont.name} (${cont.revision})${cont.latest ? ' [LATEST]' : ''}`,
      value: cont
    };
  });

  const prompt = await inquirer
    .prompt([
      {
        type: 'autocomplete',
        name: 'instance',
        message: 'Select a task and container (type to filter)',
        source: (_: any, input: string): Promise<TaskEntry[]> => {
          input = input || "";
          return new Promise((res): void => {
            const results = fuzzy.filter<TaskEntry>(input, taskList, { extract: (e): string => e.name });
            return res(results.map((e): TaskEntry => e.original));
          });
        }
      }]);
  return prompt.instance;
}

function markLatest(defs: TaskDefinitionSummary[]): TaskDefinitionSummary[] {
  const latests: Record<string, TaskDefinitionSummary> = {};
  defs.forEach((def): void => {
    const current = latests[def.name];
    if (!current || def.revision > current.revision) {
      latests[def.name] = def;
    }
  });
  for (const entry in latests) {
    if (Object.prototype.hasOwnProperty.call(latests, entry)) {
      const element = latests[entry];
      element.latest = true;
    }
  }
  return defs;
}

function extractDetails(arn: string): TaskDefinitionSummary {
  const def = arn.split('/')[1].split(':');
  return {
    arn,
    name: def[0],
    revision: parseInt(def[1], 10)
  };
}