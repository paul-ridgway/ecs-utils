import { DescribeClustersCommand, DescribeTaskDefinitionCommand, DescribeTasksCommand, ListClustersCommand, ListTasksCommand } from "@aws-sdk/client-ecs";
import { getClient } from "../utils/ecs-client-provider";
import { info } from "../utils/logger";
import inquirer from "inquirer";
import fuzzy from 'fuzzy';

export declare type TaskContainer = { taskArn: string, container: string};
export declare type TaskEntry = { name: string, value: TaskContainer; };

export async function getTask(clusterArn: string): Promise<TaskContainer> {
  info("Fetching task list...");
  const cmd = new ListTasksCommand({ cluster: clusterArn });
  const tasks = await getClient().send(cmd);
  const taskDetails = await getClient().send(new DescribeTasksCommand({ cluster: clusterArn, tasks: tasks.taskArns }));

  const taskList: TaskEntry[] = (taskDetails.tasks|| []).map((task): TaskEntry[] => (task.containers || []).map((cont): TaskEntry => {
    return {
      name: `Task Definition: ${task.taskDefinitionArn?.split("/")[1]}, Container: ${cont.name} (${cont.lastStatus})`,
      value: {
        taskArn: task.taskArn!,
        container: cont.name!
      }
    };
  })).flat();

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
