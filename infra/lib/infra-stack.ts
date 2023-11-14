import * as cdk from 'aws-cdk-lib';
import { Construct } from 'constructs';
// import * as sqs from 'aws-cdk-lib/aws-sqs';
import * as ec2 from 'aws-cdk-lib/aws-ec2';

import * as ecr from 'aws-cdk-lib/aws-ecr';
import * as ecs from 'aws-cdk-lib/aws-ecs';
import * as ecsp from 'aws-cdk-lib/aws-ecs-patterns';

export class InfraStack extends cdk.Stack {
  constructor(scope: Construct, id: string, props?: cdk.StackProps) {
    super(scope, id, props);
    const ecrRepo = new ecr.Repository(this, "GoAPI", {
      repositoryName: "go-browse-together-repo",
  });
  new cdk.CfnOutput(this, "ecrRepoURI", {
      value: ecrRepo.repositoryUri,
  });

    const loadBalancedFargateService = new ecsp.ApplicationLoadBalancedFargateService(this, 'Service', {
      memoryLimitMiB: 512,
      cpu: 256,
      taskImageOptions: {
        image: ecs.ContainerImage.fromEcrRepository(ecrRepo),
      },
    });
    
    loadBalancedFargateService.targetGroup.configureHealthCheck({
      path: '/health',
    });
  }
}
