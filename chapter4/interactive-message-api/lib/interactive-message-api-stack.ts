import cdk = require('@aws-cdk/core');
import * as iam from '@aws-cdk/aws-iam';
import { Function, Runtime, Code } from "@aws-cdk/aws-lambda"
import { RestApi, Integration, LambdaIntegration, Resource,
  MockIntegration, PassthroughBehavior, EmptyModel } from "@aws-cdk/aws-apigateway"

export class InteractiveMessageApiStack extends cdk.Stack {
  constructor(scope: cdk.Construct, id: string, props?: cdk.StackProps) {
    super(scope, id, props);

    // The code that defines your stack goes here
    // Lambda Function 作成
    const lambdaFunction: Function = new Function(this, "sample-interactive-message-lambda", {
      functionName: "sample-interactive-message-lambda", // 関数名
      runtime: Runtime.GO_1_X, // ランタイムの指定
      code: Code.asset("./lambdaSource"), // ソースコードのディレクトリ
      handler: "main", // handler の指定
      memorySize: 256, // メモリーの指定
      timeout: cdk.Duration.seconds(10), // タイムアウト時間
      environment: {
        "BOT_TOKEN": "xxxxxxxxxxxxxxxxxxxxxx"
      } // 環境変数
    })

    //Policyを関数に付加
    lambdaFunction.addToRolePolicy(new iam.PolicyStatement({
      resources: ["*"],
      actions: ["ec2:StartInstances", "ec2:StopInstances", "ec2:DescribeInstances"],
    }))

    // API Gateway 作成
    const restApi: RestApi = new RestApi(this, "sample-interactive-message-api", {
      restApiName: "sample-interactive-message-api", // API名
      description: "Deployed by CDK" // 説明
    })

    // Integration 作成
    const integration: Integration = new LambdaIntegration(lambdaFunction,
        {
          proxy: true,
          integrationResponses: [
            {
              statusCode: '200',
              responseTemplates: {
                'application/json': '$input.json("$")'
              }
            }
          ],
          requestTemplates: {
            'application/json': '$input.json("$")'
          },
        })

    // リソースの作成
    const getResouse: Resource = restApi.root.addResource("event")

    // メソッドの作成
    getResouse.addMethod("POST", integration, {methodResponses: [{statusCode: '200',}]})
    getResouse.addMethod("GET", integration)


    // CORS対策でOPTIONSメソッドを作成
    getResouse.addMethod("OPTIONS", new MockIntegration({
      integrationResponses: [{
        statusCode: "200",
        responseParameters: {
          "method.response.header.Access-Control-Allow-Headers":
              "'Content-Type,X-Amz-Date,Authorization,X-Api-Key,X-Amz-Security-Token,X-Amz-User-Agent'",
          "method.response.header.Access-Control-Allow-Origin": "'*'",
          "method.response.header.Access-Control-Allow-Credentials": "'false'",
          "method.response.header.Access-Control-Allow-Methods": "'OPTIONS,GET,PUT,POST,DELETE'",
        }
      }],
      passthroughBehavior: PassthroughBehavior.NEVER,
      requestTemplates: {
        "application/json": "{\"statusCode\": 200}"
      }
    }), {
      methodResponses: [{
        statusCode: "200",
        responseParameters: {
          "method.response.header.Access-Control-Allow-Headers": true,
          "method.response.header.Access-Control-Allow-Origin": true,
          "method.response.header.Access-Control-Allow-Credentials": true,
          "method.response.header.Access-Control-Allow-Methods": true,
        },
        responseModels: {
          "application/json": new EmptyModel()
        },
      }]
    })
  }
}
