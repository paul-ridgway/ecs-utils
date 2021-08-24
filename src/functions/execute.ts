import { ExecuteCommandCommand } from "@aws-sdk/client-ecs";
import { spawn } from "child_process";
import { getClient } from "../utils/ecs-client-provider";
import { error, info } from "../utils/logger";
import { Action } from "./action";

export class Execute extends Action {

  async run(): Promise<void> {
    const ecs = getClient();
    const cluster = await this.getClusterArn();
    const task = await this.getTask();
    // TODO: Ask command?
    // TODO: Native shell in future using https://stackoverflow.com/questions/64195313/how-to-send-commands-to-aws-session-manager-websocket-url-using-xterm-js?
    // TODO: Pass through region
    const args = ['ecs', 'execute-command', '--cluster', cluster, '--task', task.taskArn, '--container', task.container, '--command', '/bin/bash', '--interactive'];
    // For now this is a quick win
    // TODO: Check for AWS CLI, version and SSM plugin
    return new Promise<void>((res, rej): void => {
      const shell = spawn('aws', args, { stdio: 'inherit' });
      shell.on('error', (err: Error): void => { error("Shell error: " + err.message); rej(err); });
      shell.on('close', (code: number): void => { info(`Process terminated with code: ${code}`); res(); });
    });
  }

}