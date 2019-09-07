#!/usr/bin/env node
import 'source-map-support/register';
import cdk = require('@aws-cdk/core');
import { IncomingWebhookAppStack } from '../lib/incoming-webhook-app-stack';
import { SampleEc2 } from "../lib/sample_ec2"



const util = require('util');
const exec = util.promisify(require('child_process').exec);

async function deploy(){
    await exec('go get -v -t -d ./lambdaSource/... && GOOS=linux GOARCH=amd64 go build -o ./lambdaSource/main ./lambdaSource/**.go')
    const app = new cdk.App();
    new SampleEc2(app,"SampleEc2Stack")
    new IncomingWebhookAppStack(app, 'IncomingWebhookAppStack');
    app.synth()
    await  exec('rm ./lambdaSource/main')
}

deploy()
