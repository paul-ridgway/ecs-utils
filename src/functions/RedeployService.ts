import { ListTaskDefinitionsCommand, UpdateServiceCommand } from "@aws-sdk/client-ecs";
import { getClient } from "../utils/ecs-client-provider";
import { error, info } from "../utils/logger";
import { Action } from "./Action";

export class RedeployService extends Action {

  async run(): Promise<void> {
    const ecs = getClient();
    const cluster = await this.getClusterArn();
    const service = await this.getService();
    new UpdateServiceCommand({service: service, taskDefinition: taskDefinition})
    return Promise.resolve();
  }

}