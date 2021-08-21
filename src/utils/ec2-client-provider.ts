import { EC2Client } from "@aws-sdk/client-ec2";

let client: EC2Client;

export function getClient(): EC2Client {
  if (!client) {
    client = new EC2Client({ region: "eu-west-2" });
  }
  return client;
}