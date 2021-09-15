import inquirer from "inquirer";
import fuzzy from 'fuzzy';
import { ListClustersCommand } from "@aws-sdk/client-ecs";
import { info } from "../utils/logger";
import { getClient } from "../utils/ecs-client-provider";

declare type ClusterEntry = { name: string, value: string; };
export async function getCluster(): Promise<string> {
  info("Fetching cluster list...");
  const cmd = new ListClustersCommand({ maxResults: 100 });
  const clusters = await getClient().send(cmd);
  const clusterList: ClusterEntry[] = clusters.clusterArns?.map((arn): ClusterEntry => ({ name: arn.split('/')[1], value: arn })) || [];
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
