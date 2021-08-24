import { ExecuteCommandCommand } from "@aws-sdk/client-ecs";
import { spawn } from "child_process";
import { getClient } from "../utils/ecs-client-provider";
import { Action } from "./action";

export class Execute extends Action {

  async run() {
    const ecs = getClient();
    const cluster = await this.getClusterArn();
    const task = await this.getTask();
    // TODO: Ask command?
    // TODO: Native shell in future using https://stackoverflow.com/questions/64195313/how-to-send-commands-to-aws-session-manager-websocket-url-using-xterm-js?
    // TODO: Pass through region
    const args = ['ecs', 'execute-command', '--cluster', cluster, '--task', task.taskArn, '--container', task.container, '--command', '/bin/bash', '--interactive'];
    // For now this is a quick win
    // TODO: Check for AWS CLI, version and SSM plugin
    return new Promise<void>((res, rej) => {
      const shell = spawn('aws', args, { stdio: 'inherit' });
      shell.on('error', (err: Error) => { console.error(err.message); rej(err); });
      shell.on('close', (code: number) => { console.log('[shell] terminated :', code); res(); });
    });
  }

}