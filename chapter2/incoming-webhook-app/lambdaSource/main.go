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
	"github.com/pkg/errors"
	"os"
)

type EC2Status struct {
	InstanceID string `json:"instance-id"`
	State      string `json:"state"`
}

type EC2_STATE string

const (
	EC2_RUNNING_STATE EC2_STATE = "running"
	EC2_STOPPED_STATE EC2_STATE = "stopped"
)

const (
	SLACK_ICON = ":ok:"
	SLACK_NAME = "Sample Notice"
)

const EC2_RESOURCE = "aws.ec2"

func main() {
	lambda.Start(noticeHandler)
}

func noticeHandler(context context.Context,
	event events.CloudWatchEvent) (e error) {
	webHookURL := os.Getenv("webHookUrl")
	channel := os.Getenv("slackChannel")
	switch event.Source {
	case EC2_RESOURCE:
		if err := notifyEC2Status(event, webHookURL, channel); err != nil {
			log.Error(err)
			return err
		}
		return nil
	default:
		return errors.New("想定するリソースのイベントではない")
	}
}

func notifyEC2Status(event events.CloudWatchEvent,
	webHookURL string, channel string) (e error) {
	status := &EC2Status{}
	err := json.Unmarshal([]byte(event.Detail), status)
	if err != nil {
		log.Error(err)
		return err
	}

	ec := ec2.New(session.New(), &aws.Config{Region: aws.String("ap-northeast-1")})
	descOutput, err := ec.DescribeInstances(&ec2.DescribeInstancesInput{
		InstanceIds: []*string{&status.InstanceID},
	})

	var instanceName string
	for _, tag := range descOutput.Reservations[0].Instances[0].Tags {
		if *tag.Key == "Name" {
			instanceName = *tag.Value
		}
	}

	var title string
	switch EC2_STATE(status.State) {
	case EC2_RUNNING_STATE:
		title = fmt.Sprintf("*%vが%v*", instanceName, status.State)
	case EC2_STOPPED_STATE:
		title = fmt.Sprintf("*%vが%v*", instanceName, status.State)
	default:
		return nil
	}

	err = ReportToSlack(webHookURL, SlackRequestBody{
		IconEmoji: aws.String(SLACK_ICON),
		Username:  SLACK_NAME,
		Channel:   channel,
		Text:      title,
	})

	return nil
}
