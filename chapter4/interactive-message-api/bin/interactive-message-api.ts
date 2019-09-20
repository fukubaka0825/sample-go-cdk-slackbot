#!/usr/bin/env node
import 'source-map-support/register';
import cdk = require('@aws-cdk/core');
import { Interactive_message_api_stack } from '../lib/interactive-message-api-stack';

const util = require('util');
const exec = util.promisify(require('child_process').exec);

async function deploy(){

    await exec('go get -v -t -d ./lambdaSource/... && ' +
        'GOOS=linux GOARCH=amd64 go build -o ./lambdaSource/main ./lambdaSource/**.go');

    const app = new cdk.App();
    new Interactive_message_api_stack(app, 'InteractiveMessageApiStack');
    app.synth();

    await exec('rm ./lambdaSource/main');
}

deploy()

