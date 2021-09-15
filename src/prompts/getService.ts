import { DescribeServicesCommand, DescribeTasksCommand, ListServicesCommand, ListTasksCommand, Service } from "@aws-sdk/client-ecs";
import { getClient } from "../utils/ecs-client-provider";
import { info } from "../utils/logger";
import inquirer from "inquirer";
import fuzzy from 'fuzzy';

export declare type ServiceEntry = { name: string, value: Service; };

export async function getService(clusterArn: string): Promise<Service> {
  info("Fetching service list...");
  const cmd = new ListServicesCommand({ cluster: clusterArn });
  const services = await getClient().send(cmd);
  const serviceDetails = await getClient().send(new DescribeServicesCommand({ cluster: clusterArn, services: services.serviceArns }));

  const taskList: ServiceEntry[] = (serviceDetails.services|| []).map((service): ServiceEntry => {
    return {
      name: `Name: ${service.serviceName}, Task Definition: ${service.taskDefinition!.split('/')[1]}`,
      value: service,
    };
  });;

  const prompt = await inquirer
    .prompt([
      {
        type: 'autocomplete',
        name: 'instance',
        message: 'Select a service (type to filter)',
        source: (_: any, input: string): Promise<ServiceEntry[]> => {
          input = input || "";
          return new Promise((res): void => {
            const results = fuzzy.filter<ServiceEntry>(input, taskList, { extract: (e): string => e.name });
            return res(results.map((e): ServiceEntry => e.original));
          });
        }
      }]);
  return prompt.instance;
}
