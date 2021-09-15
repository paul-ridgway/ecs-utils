import { Service } from "@aws-sdk/client-ecs";
import { getCluster } from "../prompts/getCluster";
import { getService } from "../prompts/get";
import { getTask, TaskContainer } from "../prompts/get-task";

export abstract class Action {

  private _clusterArn: string = '';

  public constructor() {

  }

  public async getClusterArn(): Promise<string> {
    if (!this._clusterArn) {
      this._clusterArn = await getCluster();
    }

    return this._clusterArn;
  }

  public async getService(): Promise<Service> {
    return getService(await this.getClusterArn());
  }

  public async getTask(): Promise<TaskContainer> {
    return getTask(await this.getClusterArn());
  }

  public async getTaskDefinition(): Promise<string> {
    return getTaskDefinition(await this.getClusterArn())
  }

  abstract run(): Promise<void>;

}
