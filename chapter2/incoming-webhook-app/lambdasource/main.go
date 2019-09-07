package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/labstack/gommon/log"
	"os"
)

type EC2Status struct {
	InstanceID    string  `json:"instance-id"`
	State         string  `json:"state"`
}


type EC2_STATE string

const (
	EC2_RUNNING_STATE   EC2_STATE = "running"
	EC2_STOPPED_STATE    EC2_STATE = "stopped"
)

const (
	SLACK_ICON = ":ok:"
	SLACK_NAME = "Sample Notice"
)

func main() {
	lambda.Start(noticeHandler)
}

func noticeHandler(context context.Context, event events.CloudWatchEvent) (e error) {
	webhookURL := os.Getenv("webHookUrl")
	channel := os.Getenv("slackChannel")
	switch event.Source {
	case "aws.ec2":
		if err := notifyEC2Status(event, webhookURL, channel); err != nil {
			log.Error(err)
			return err
		}
		return nil
	default:
		log.Info("想定するリソースのイベントではない")
		return nil
	}
}

func notifyEC2Status(event events.CloudWatchEvent, webhookURL string, channel string) (e error) {
	status := &EC2Status{}
	err := json.Unmarshal([]byte(event.Detail), status)
	if err != nil {
		log.Error(err)
		return err
	}

	ec := ec2.New(session.New(), &aws.Config{Region: aws.String("ap-northeast-1")})
	descOutput,err := ec.DescribeInstances(&ec2.DescribeInstancesInput{
		InstanceIds:[]*string{&status.InstanceID},
	})


	var instanceName string
	for _,tag := range descOutput.Reservations[0].Instances[0].Tags{
		if *tag.Key == "Name"{
			instanceName = *tag.Value
		}
	}

	text := fmt.Sprintf("*instanceName:%v* \n *state:%v*",instanceName,status.State)

	var messsageLevel SLACK_MESSAGE_LEVEL = SLACK_MESSAGE_LEVEL_OK
	var title string
	switch EC2_STATE(status.State) {
	case EC2_RUNNING_STATE:
		title = fmt.Sprintf("*%vが%v*", instanceName, status.State)
		messsageLevel = SLACK_MESSAGE_LEVEL_NOTIFY

	case EC2_STOPPED_STATE:
		title = fmt.Sprintf("*%vが%v*", instanceName,status.State)
		messsageLevel = SLACK_MESSAGE_LEVEL_ALART

	default:
		return nil
	}

	err = ReportToLaboonSlack(webhookURL, channel, SLACK_NAME, SLACK_ICON, title, text, messsageLevel)
	if err != nil {
		log.Error(err)
		return err
	}
	return nil
}

