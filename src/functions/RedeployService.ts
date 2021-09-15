import { ListTaskDefinitionsCommand, UpdateServiceCommand } from "@aws-sdk/client-ecs";
import { getClient } from "../utils/ecs-client-provider";
import { error, info } from "../utils/logger";
import { Action } from "./Action";

export class RedeployService extends Action {

  async run(): Promise<void> {
    const cluster = await this.getClusterArn();
    const service = await this.getService();
    const taskDefinition = await this.getTaskDefinition(extractTaskDefinitionFamily(service.taskDefinition));
    // TODO: Ask to 'just deploy latest'
    // TODO: Ask force deployment
    const cmd = new UpdateServiceCommand({cluster, service: service.serviceArn!, taskDefinition: taskDefinition.arn, forceNewDeployment: true})
    const res = await getClient().send(cmd)
    info("Service update command sent successfully");

    // TODO: Monitor for deployment done
    return Promise.resolve();
  }

}

function extractTaskDefinitionFamily(arn?: string): string | undefined {
  if (!arn) {
    return;
  }
  return arn.split('/')[1].split(':')[0];
}