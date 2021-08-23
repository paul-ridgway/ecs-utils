import { ExecuteCommandCommand } from "@aws-sdk/client-ecs";
import { getClient } from "../utils/ecs-client-provider";
import { Action } from "./action";

export class Execute extends Action {
  
  async run() {
    const ecs = getClient();
    const params = {
      command: 'STRING_VALUE', /* required */
      interactive: true || false, /* required */
      task: 'STRING_VALUE', /* required */
      cluster: 'STRING_VALUE',
      container: 'STRING_VALUE'
    };
    const cmd = new ExecuteCommandCommand({
      cluster: await this.getClusterArn(),
      command: 'bash',
      task: await this.getTask(),
      interactive: true,
    })
    await ecs.send(cmd);
  }

}