import { spawn } from "child_process";
import inquirer from "inquirer";
import { error, info, warn } from "../utils/logger";
import { Action } from "./Action";

export class Execute extends Action {

  async run(): Promise<void> {
    const cluster = await this.getClusterArn();
    const task = await this.getTask();
    const askCommand = await inquirer.prompt({
      type: "input",
      name: "command",
      message: "Command to run",
      default: "/bin/sh",
    })
    // TODO: Native shell in future using https://stackoverflow.com/questions/64195313/how-to-send-commands-to-aws-session-manager-websocket-url-using-xterm-js?
    // TODO: Pass through region
    const args = ['ecs', 'execute-command', '--cluster', cluster, '--task', task.taskArn, '--container', task.container, '--command', askCommand.command, '--interactive'];
    // For now this is a quick win
    // TODO: Check for AWS CLI, version and SSM plugin

    // Trap ctrl-c and do not exit
    let last = 0;
    process.on('SIGINT', () => {
      const now = Date.now();
      if (now - last < 1000) {
        process.exit();
      }
      last = now;
    });    

    info("Note: Press ctrl-c twice within a second to exit, once to send to the container");
    await new Promise<void>((res, rej): void => {
      const shell = spawn('aws', args, { stdio: 'inherit' });
      shell.on('error', (err: Error): void => { error("Shell error: " + err.message); rej(err); });
      shell.on('close', (code: number): void => { info(`Process terminated with code: ${code}`); res(); });
    });

    // Remove trap 
    process.removeAllListeners('SIGINT');
  }

}