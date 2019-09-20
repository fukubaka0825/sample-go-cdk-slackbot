package main

import (
	"encoding/json"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/labstack/gommon/log"
	"github.com/nlopes/slack"
	"github.com/pkg/errors"
	"net/http"
	"net/url"
	"strings"
)

const (
	ACTION_SELECT = "select"
	ACTION_START  = "start"
	ACTION_CANCEL = "cancel"
)

type InteractiveMessageUsecase interface {
	MakeSlackResponse(request events.APIGatewayProxyRequest) (events.
		APIGatewayProxyResponse, error)
}

func NewInteractionUsecase(signingSecrets string) InteractiveMessageUsecase {
	return &interactiveMessageUsecase{
		signingSecrets: signingSecrets,
	}
}

type interactiveMessageUsecase struct {
	signingSecrets string
}

func (i *interactiveMessageUsecase) MakeSlackResponse(request events.APIGatewayProxyRequest) (events.
	APIGatewayProxyResponse, error) {
	response := events.APIGatewayProxyResponse{}

	str, _ := url.QueryUnescape(request.Body)
	str = strings.Replace(str, "payload=", "", 1)
	log.Infof("str:%v type:%T", str, str)

	var message slack.InteractionCallback
	if err := json.Unmarshal([]byte(str), &message); err != nil {
		return events.APIGatewayProxyResponse{Body: "json error", StatusCode: 500}, nil
	}

	if request.HTTPMethod != http.MethodPost {
		response.StatusCode = http.StatusMethodNotAllowed
		return response, errors.New("Invalid method")
	}

	if err := i.verify(request); err != nil {
		log.Error(err)
		return response, err
	}

	action := message.ActionCallback.AttachmentActions[0]
	switch action.Name {
	case ACTION_SELECT:
		value := action.SelectedOptions[0].Value

		originalMessage := message.OriginalMessage

		originalMessage.Attachments[0].Text = fmt.
			Sprintf("%s を本当に %s してしまって良いか？ ",
				strings.Split(value, ":")[0], strings.Split(value, ":")[1])

		originalMessage.Attachments[0].Actions = []slack.AttachmentAction{
			{
				Name:  string(ACTION_START),
				Text:  "はい",
				Type:  "button",
				Value: value,
				Style: "primary",
			},
			{
				Name:  string(ACTION_CANCEL),
				Text:  "やっぱいいや",
				Type:  "button",
				Style: "danger",
			},
		}

		resJson, err := json.Marshal(&originalMessage)
		if err != nil {
			return response, err
		}

		response.Body = string(resJson)
		response.Headers = make(map[string]string)
		response.Headers["Content-Type"] = "application/json"
		response.StatusCode = http.StatusOK

		return response, nil
	case ACTION_START:
		value := action.Value
		title := fmt.Sprintf("%sを開始しました！", value)

		ec := ec2.New(session.New(), &aws.Config{
			Region: aws.String("ap-northeast-1")})

		descOutput, err := ec.DescribeInstances(&ec2.DescribeInstancesInput{
			Filters: []*ec2.Filter{
				&ec2.Filter{
					Name: aws.String("tag:Name"),
					Values: []*string{
						aws.String(strings.Split(value, ":")[0]),
					},
				},
			},
		})
		if err != nil {
			return response, err
		}

		if strings.Split(value, ":")[1] == "停止" {
			_, err = ec.StopInstances(&ec2.StopInstancesInput{
				InstanceIds: []*string{
					descOutput.Reservations[0].Instances[0].InstanceId,
				},
			})
		}

		if strings.Split(value, ":")[1] == "再起動" {
			_, err = ec.StartInstances(&ec2.StartInstancesInput{
				InstanceIds: []*string{
					descOutput.Reservations[0].Instances[0].InstanceId,
				},
			})
		}

		if err != nil {
			return response, errors.New("インスタンスの再起動/停止に失敗")
		}

		return makeResponse(&response, message.OriginalMessage, title, "")
	case ACTION_CANCEL:
		title := fmt.Sprintf(":x: @%s はサーバーの状態変更を辞めたようだ",
			message.User.Name)
		return makeResponse(&response, message.OriginalMessage, title, "")
	default:
		response.StatusCode = http.StatusInternalServerError
		return response, errors.New("Invalid action was submitted")
	}
}

func (i *interactiveMessageUsecase) verify(request events.
	APIGatewayProxyRequest) error {
	httpHeader := http.Header{}
	for key, value := range request.Headers {
		httpHeader.Set(key, value)
	}
	sv, err := slack.NewSecretsVerifier(httpHeader, i.signingSecrets)
	if err != nil {
		log.Error(err)
		return err
	}

	if _, err := sv.Write([]byte(request.Body)); err != nil {
		log.Error(err)
		return err
	}

	if err := sv.Ensure(); err != nil {
		log.Error("Invalid SIGNING_SECRETS")
		return err
	}
	return nil
}

//ボタンを空にしてresponseを作成
func makeResponse(response *events.APIGatewayProxyResponse,
	original slack.Message, title, value string) (events.APIGatewayProxyResponse, error) {
	if original.Attachments == nil {
		original.Attachments = []slack.Attachment{slack.Attachment{}}
	}

	original.Attachments[0].Actions = []slack.AttachmentAction{}
	original.Attachments[0].Fields = []slack.AttachmentField{
		{
			Title: title,
			Value: value,
			Short: false,
		},
	}
	resJson, err := json.Marshal(&original)
	if err != nil {
		return *response, errors.New("originalMessageのmarshalに失敗")
	}
	response.Body = string(resJson)
	response.Headers = make(map[string]string)
	response.Headers["Content-type"] = "application/json"
	response.StatusCode = http.StatusOK
	return *response, nil
}
