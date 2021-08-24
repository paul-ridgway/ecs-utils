import { TaskContainer } from "../prompts/get-task";

type ClusterFn = () => Promise<string>;
type TaskFn = (clusterArn: string) => Promise<TaskContainer>;

export abstract class Action {

  private _clusterArn: string = '';

  public constructor(private _getClusterArn: ClusterFn, private _getServiceId: ClusterFn, private _getTaskArn: TaskFn) {

  }

  public async getClusterArn(): Promise<string> {
    if (!this._clusterArn) {
      this._clusterArn = await this._getClusterArn();
    }

    return this._clusterArn;
  }

  public getServiceId(): Promise<string> {
    return this._getServiceId();
  }

  public async getTask(): Promise<TaskContainer> {
    return this._getTaskArn(await this.getClusterArn())
  }

  abstract run(): Promise<void>;

}