import cdk = require('@aws-cdk/core');
import * as iam from '@aws-cdk/aws-iam';
import { Function, Runtime, Code } from "@aws-cdk/aws-lambda"
import { RestApi, Integration, LambdaIntegration, Resource,
  MockIntegration, PassthroughBehavior, EmptyModel } from "@aws-cdk/aws-apigateway"

export class Interactive_message_api_stack extends cdk.Stack {
  constructor(scope: cdk.Construct, id: string, props?: cdk.StackProps) {
    super(scope, id, props);
    
    // Lambda Function 作成
    const lambdaFunction: Function = new Function(this, "SampleInteractiveMessageLambda", {
      functionName: "sample-interactive-message-lambda", 
      runtime: Runtime.GO_1_X, 
      code: Code.asset("./lambdaSource"), 
      handler: "main", 
      memorySize: 256, 
      timeout: cdk.Duration.seconds(10), 
      environment: {
        "SIGNING_SECRETS":"fd9375aab367axxxxxxxxxxxxxxxxx",
      } 
    })

    //Policyを関数に付加
    lambdaFunction.addToRolePolicy(new iam.PolicyStatement({
      resources: ["*"],
      actions: ["ec2:StartInstances", "ec2:StopInstances", "ec2:DescribeInstances"],
    }))

    // API Gateway 作成
    const restApi: RestApi = new RestApi(this, "SampleInteractiveMessageApi", {
      restApiName: "sample-interactive-message-api", // API名
      description: "Deployed by CDK" // 説明
    })

    // Integration 作成
    const integration: Integration = new LambdaIntegration(lambdaFunction)

    // リソースの作成
    const getResouse: Resource = restApi.root.addResource("event")

    // メソッドの作成
    getResouse.addMethod("POST", integration)

  }
}
