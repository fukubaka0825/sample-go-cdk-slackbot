package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/nlopes/slack"
	"github.com/labstack/gommon/log"
	"net/http"
	"net/url"
	"strings"
	"syscall"
)

const (
	actionSelect   = "select"
	actionCancel   = "cancel"
)

type ApiEvent struct{
	Type       string   `json:"type"`
	Text       string   `json:"text"`
	Challenge  string   `json:"challenge"`
	Token      string   `json:"token"`
	Event      Event    `json:"event"`
}
type Event struct	{
	User  string `json:"user"`
	Type  string `json:"type"`
	Text  string `json:"text"`
	Channel string `json:"channel"`
}

func main() {
	lambda.Start(handleRequest)
}

// TODO:雑すぎるのであとでリファクタ
func handleRequest(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {

	log.Info("開始")
	vals, _ := url.ParseQuery(request.Body)
	log.Infof("vals:%v type:%T",vals,vals)
	response := events.APIGatewayProxyResponse{}

	//環境変数チェック
	botToken, found := syscall.Getenv("BOT_TOKEN")
	if !found {
		log.Info("Token Not Found")
		return response, errors.New("Token Not Found")
	}
	channelID, found := syscall.Getenv("CHANNEL_ID")
	if !found {
		log.Info("Channel Id Not Found")
		return response,errors.New("Channel Id Not Found")}

	botID, found := syscall.Getenv("BOT_ID")
	if !found {
		log.Info("Bot Id Not Found")
		return response,errors.New("Bot Id Not Found")
	}
	botOAuth, found := syscall.Getenv("BOT_OAUTH")
	if !found {
		log.Info("OAuth Access Token Not Found")
		return response, errors.New("OAuth Access Token Not Found")
	}


	apiEvent := &ApiEvent{}
	for key,_ := range vals{
		err := json.Unmarshal([]byte(key),apiEvent)
		if err != nil{
			return response,err
		}
	}
	switch apiEvent.Type {
	case "url_verification":
		log.Info("url_verification")
		//mapは初期化されないっぽい
		response.Headers = make(map[string]string)
		response.Headers["Content-Type"] = "text/plain"
		response.Body = apiEvent.Challenge
		response.StatusCode = http.StatusOK
		return response, nil
	case "event_callback":
		client := slack.New(botOAuth)

		event := apiEvent.Event

		//input validate
		if event.Type != "message" || !strings.HasPrefix(event.Text, fmt.Sprintf("<@%s> ", botID)) {
			log.Infof("%s %s",apiEvent.Event.User ,botID)
			return response,errors.New("bot Id Not Found")
		}

		if event.Channel != channelID {
			log.Infof("%s %s",apiEvent.Event.Channel , channelID)
			return response,errors.New("channel Id Not Found")
		}
		if apiEvent.Token != botToken {
			log.Infof("%s %s", apiEvent.Token, botToken)
			return response,errors.New("botToken Not Found")
		}
		m := strings.Split(strings.TrimSpace(event.Text), " ")[1:]
		if len(m) == 0 ||(m[0] != "down" && m[0] != "up") {
			return response, fmt.Errorf("invalid message")
		}


		ec := ec2.New(session.New(), &aws.Config{Region: aws.String("ap-northeast-1")})

		var(
			operation string
			targetEc2Status string
		)
		if m[0] == "down" {
			operation = "停止"
			targetEc2Status = "running"
		}
		if m[0] == "up" {
			operation = "再起動"
			targetEc2Status = "stopped"
		}

		descOutput,err := ec.DescribeInstances(&ec2.DescribeInstancesInput{
			Filters:[]*ec2.Filter{
				&ec2.Filter{
					Name:aws.String("instance-state-name"),
					Values:[]*string{
						aws.String(targetEc2Status),
					},
				},
			},
		})

		if err != nil{
			return response,err
		}


		options := []slack.AttachmentActionOption{}
		for _,reservation := range descOutput.Reservations{
			for _,instance := range reservation.Instances{
				var instanceName string
				for _,tag := range instance.Tags{
					if *tag.Key=="Name"{
						instanceName = *tag.Value
					}
				}
				options = append(options,slack.AttachmentActionOption{
					Text:instanceName,
					Value:instanceName+":"+operation,
				})
			}
		}

		attachment := slack.Attachment{
			Text:       "どのサーバーが"+operation+"対象かの? :server:",
			Color:      "#f9a41b",
			CallbackID: "server",
			Actions: []slack.AttachmentAction{
				{
					Name: actionSelect,
					Type: "select",
					Options:options,
				},

				{
					Name:  actionCancel,
					Text:  "やっぱ何でもないわ",
					Type:  "button",
					Style: "danger",
				},
			},
		}

		params := slack.PostMessageParameters{
			IconEmoji: ":server:",
			Username:  "Server Sample",
		}


		msgOptText := slack.MsgOptionText("", true)
		msgOptParams := slack.MsgOptionPostMessageParameters(params)
		msgOptAttachment := slack.MsgOptionAttachments(attachment)

		if _, _, err := client.PostMessage(channelID,msgOptText, msgOptParams,msgOptAttachment); err != nil {
			return response,fmt.Errorf("failed to post message: %s", err)
		}

		response.StatusCode = http.StatusOK
		return response,nil
	default:
		response.StatusCode = http.StatusOK
		return response,nil
	}

}
