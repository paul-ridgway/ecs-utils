import { ECSClient } from "@aws-sdk/client-ecs";

let client: ECSClient;

export function getClient(): ECSClient {
  if (!client) {
    // TODO: better selection/support of region
    client = new ECSClient({ region: "eu-west-2" });
  }
  return client;
}