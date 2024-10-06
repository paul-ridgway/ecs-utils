import { ListTaskDefinitionsCommand, UpdateServiceCommand } from "@aws-sdk/client-ecs";
import { getClient } from "../utils/ecs-client-provider";
import { error, info } from "../utils/logger";
import { Action } from "./Action";
import inquirer from "inquirer";

export class ScaleService extends Action {

  async run(): Promise<void> {
    const cluster = await this.getClusterArn();
    const service = await this.getService();

    info(`Currently, pending: ${service.pendingCount}, running: ${service.runningCount} of desired: ${service.desiredCount}`);
    // Ask count
    const { count } = await inquirer.prompt({
      type: "number",
      name: "count",
      message: "Desired count",
      default: service.desiredCount,
    })

    const cmd = new UpdateServiceCommand({cluster, service: service.serviceArn!, forceNewDeployment: true, desiredCount: count})
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