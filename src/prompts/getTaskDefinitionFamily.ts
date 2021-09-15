import { DescribeTaskDefinitionCommand, DescribeTaskDefinitionCommandOutput, DescribeTasksCommand, ListTaskDefinitionFamiliesCommand, ListTaskDefinitionsCommand, ListTasksCommand, TaskDefinition } from "@aws-sdk/client-ecs";
import { getClient } from "../utils/ecs-client-provider";
import { info } from "../utils/logger";
import inquirer from "inquirer";
import fuzzy from 'fuzzy';

export declare type TaskDefinitionSummary = {
  name: string,
  arn: string,
  revision: number;
};
export declare type TaskEntry = { name: string, value: string; };

export async function getTaskDefinitionFamily(familyPrefix: string | undefined): Promise<TaskDefinitionSummary> {
  info("Fetching task definition list...");
  const client = getClient();
  // TODO: Handle pagination in reponses
  const taskDefinitionFamilies = await client.send(new ListTaskDefinitionFamiliesCommand({ }));


  const taskList: TaskEntry[] = (taskDefinitionFamilies.families || []).map((family): TaskEntry => {
    return {
      name: family,
      value: family
    };
  });

  const prompt = await inquirer
    .prompt([
      {
        type: 'autocomplete',
        name: 'instance',
        message: 'Select a Task Definition Family (type to filter)',
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
