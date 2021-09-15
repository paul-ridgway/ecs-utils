import { Service } from "@aws-sdk/client-ecs";
import { getCluster } from "../prompts/getCluster";
import { getService } from "../prompts/getService";
import { getTask, TaskContainer } from "../prompts/getTask";
import { getTaskDefinition, TaskDefinitionSummary } from "../prompts/getTaskDefinition";

export abstract class Action {

  private _clusterArn: string = '';

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

  public async getTaskDefinition(familyPrefix?: string): Promise<TaskDefinitionSummary> {
    return getTaskDefinition(familyPrefix);
  }

  abstract run(): Promise<void>;

}
