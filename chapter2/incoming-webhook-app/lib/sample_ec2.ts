import cdk = require("@aws-cdk/core")
import ec2 = require("@aws-cdk/aws-ec2")
import {Vpc} from "@aws-cdk/aws-ec2"
import {CfnTag} from "@aws-cdk/core/lib/cfn-tag";
export class SampleEc2 extends cdk.Stack {
    constructor(scope: cdk.App, id: string, props?: cdk.StackProps) {
        super(scope, id, props)
        const vpc = new Vpc(this, 'ExampleVpc', {
            cidr: '10.0.0.0/16'
        })
        for(let i = 0; i < 3 ; i++){
            new ec2.CfnInstance(this, "cdktest-"+i, {
                imageId: 'ami-0f9ae750e8274075b',
                instanceType: 't2.micro',
                subnetId: vpc.publicSubnets[0].subnetId,
                securityGroupIds: [vpc.vpcDefaultSecurityGroup],
                tags:[
                    {
                        key:"Name",
                        value:"cdktest-"+i,
                    },
                    {
                        key:"Demo",
                        value:"cdktest-"+i,
                    }
                ],
            })
        }
    }
}
