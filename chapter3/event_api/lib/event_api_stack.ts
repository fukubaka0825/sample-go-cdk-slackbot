import cdk = require('@aws-cdk/core');
import { Function, Runtime, Code } from "@aws-cdk/aws-lambda"
import { RestApi, Integration, LambdaIntegration, Resource
} from "@aws-cdk/aws-apigateway"
import * as iam from '@aws-cdk/aws-iam';


export class EventApiStack extends cdk.Stack {
  constructor(scope: cdk.Construct, id: string, props?: cdk.StackProps) {
    super(scope, id, props);
    
    // Lambda Function 作成
    const lambdaFunction: Function = new Function(this, "SampleEventLambda", {
      functionName: "sample-event-lambda", 
      runtime: Runtime.GO_1_X, 
      code: Code.asset("./lambdaSource"), 
      handler: "main", 
      memorySize: 256, 
      timeout: cdk.Duration.seconds(10), 
      environment: {
        "BOT_TOKEN": "FhZRpq1TuJPZoGF8ehL8lB1y",
        "CHANNEL_ID": "CMSS8AQPM",
        "BOT_ID": "UMSSPETPD",
        "BOT_OAUTH": "xoxb-740892358499-740907503795-YYIpC03i14e17sxOiNf9OXQR"
      } // 環境変数
    })

    //Policyを関数に付加
    lambdaFunction.addToRolePolicy(new iam.PolicyStatement({
      resources: ["*"],
      actions: ["ec2:DescribeInstances"],
    }))

    // API Gateway 作成
    const restApi: RestApi = new RestApi(this, "sample-event-api", {
      restApiName: "Sample-Event-API", 
      description: "Deployed by CDK" 
    })

    // Integration 作成
    const integration: Integration = new LambdaIntegration(lambdaFunction)

    // リソースの作成
    const getResouse: Resource = restApi.root.addResource("event")

    // メソッドの作成
    getResouse.addMethod("POST", integration)
  }
}
