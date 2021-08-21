#!/usr/bin/env node

import inquirer from "inquirer";
import { info } from "./utils/logger";
import { getClient } from "./utils/ec2-client-provider";
import { Action } from "./functions/functions";
import autocomplete from 'inquirer-autocomplete-prompt';
import fuzzy from 'fuzzy';
// TODO: Profile select
// TODO: Properly handle regions?

inquirer.registerPrompt('autocomplete', autocomplete);

// declare type InstanceEntry = { name: string, value: string; };

// async function selectInstance(): Promise<string> {
//   const client = getClient();
//   info("Fetching instance list...");
//   // TODO: Iterate over tokens
//   const cmd = new DescribeInstancesCommand({ MaxResults: 1000 });
//   const reservations = await client.send(cmd);
//   const instances = reservations.Reservations!.map((r): Instance[] => r.Instances ?? []).flat();
//   const statuses = await client.send(new DescribeInstanceStatusCommand({ InstanceIds: instances.filter((i): boolean => !!i.InstanceId).map((i): string => i.InstanceId!) }));
//   const instanceList: InstanceEntry[] = [];
//   instances.forEach((i): void => {
//     const tags = i.Tags ?? [];
//     const name = tags
//       .filter((t): boolean => t.Key?.toLocaleLowerCase() === "name")
//       .map((t): string => t.Value ?? "")
//       .join(' ');
//     const status = statuses.InstanceStatuses?.filter((s): boolean => s.InstanceId === i.InstanceId)[0];
//     instanceList.push({
//       name: `Instance: ${i.InstanceId} - ${name} - ${status?.SystemStatus?.Status} / ${status?.InstanceStatus?.Status}`,
//       value: i.InstanceId!
//     });
//   });

//   const prompt = await inquirer
//     .prompt([
//       {
//         type: 'autocomplete',
//         name: 'instance',
//         message: 'Select an instance (type to filter)',
//         source: (_: any, input: string): Promise<InstanceEntry[]> => {
//           input = input || "";
//           return new Promise((res): void => {
//             const results = fuzzy.filter<InstanceEntry>(input, instanceList, { extract: (e): string => e.name });
//             return res(results.map((e): InstanceEntry => e.original));
//           });
//         }
//       }]);
//   return prompt.instance;
// }

async function selectAction(): Promise<Action> {
  const prompt = await inquirer
    .prompt([{
      type: 'list',
      name: 'action',
      loop: false,
      message: 'What do you want to do?',
      choices: [
        { name: 'Execute Command', value: undefined },
      ],
    },
    ]);
  return prompt.action;
}

async function run(): Promise<void> {
  const action = await selectAction();
  // await action(instance);
}

run()
  .then((): void => process.exit(0))
  .catch((err): void => { throw (err); });
