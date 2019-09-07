package main

import (
	"encoding/json"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/labstack/gommon/log"
	"github.com/nlopes/slack"
	"github.com/pkg/errors"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"net/http"
	"net/url"
	"strings"
)


const (
	// action is used for slack attachment action.
	actionSelect   = "select"
	actionStart    = "start"
	actionCancel   = "cancel"
)

type InteractionHandler interface {
	LambdaHandle(request events.APIGatewayProxyRequest)(events.APIGatewayProxyResponse, error)
}


func NewInteractionHandler(botToken string)InteractionHandler{
	return interactionHandler{
	 	botToken:botToken,
	}
}

// interactionHandler handles interactive message response.
type interactionHandler struct {
	botToken string
}

func (h interactionHandler) LambdaHandle(request events.APIGatewayProxyRequest)(events.APIGatewayProxyResponse, error) {
	response := events.APIGatewayProxyResponse{}

	str, _ := url.QueryUnescape(request.Body)
	str = strings.Replace(str, "payload=", "", 1)

	var message slack.InteractionCallback
	if err := json.Unmarshal([]byte(str), &message); err != nil {
		return events.APIGatewayProxyResponse{Body: "json error", StatusCode: 500}, nil
	}

	if request.HTTPMethod != http.MethodPost {
		log.Printf("[ERROR] Invalid method: %s", request.HTTPMethod)
		response.StatusCode =http.StatusMethodNotAllowed
		return response, errors.New("Invalid method")
	}


	if message.Token != h.botToken {
		log.Printf("[ERROR] Invalid token: %s", message.Token)
		response.StatusCode =http.StatusUnauthorized
		return  response, errors.New("Invalid token")
	}

	action := message.ActionCallback.AttachmentActions[0]
	switch action.Name {
	case actionSelect:
		value := action.SelectedOptions[0].Value

		originalMessage := message.OriginalMessage

		originalMessage.Attachments[0].Text = fmt.Sprintf("%s を本当に %s してしまって良いか？ ",strings.Split(value,":")[0], strings.Split(value,":")[1])
		originalMessage.Attachments[0].Actions = []slack.AttachmentAction{
			{
				Name:  actionStart,
				Text:  "はい",
				Type:  "button",
				Value: value ,
				Style: "primary",
			},
			{
				Name:  actionCancel,
				Text:  "やっぱいいや",
				Type:  "button",
				Style: "danger",
			},
		}
		resJson, err := json.Marshal(&originalMessage)
		if err != nil{
			return  response, err
		}
		response.Body = string(resJson)
		response.Headers = make(map[string]string)
		response.Headers["Content-Type"] = "application/json"
		response.StatusCode =http.StatusOK
		return response,nil
	case actionStart:
		value := action.Value
		log.Infof("value:%v",value)
		title := fmt.Sprintf("%sを開始しました！", value)
		ec := ec2.New(session.New(), &aws.Config{Region: aws.String("ap-northeast-1")})
		descOutput,err := ec.DescribeInstances(&ec2.DescribeInstancesInput{
			Filters:[]*ec2.Filter{
				&ec2.Filter{
					Name:aws.String("tag:Name"),
					Values:[]*string{
						aws.String(strings.Split(value,":")[0]),
					},
				},
			},
		})
		log.Infof("descOutput:%v",descOutput)
		if err != nil {
			return response, err
		}
		if strings.Split(value,":")[1]=="停止"{
			_,err = ec.StopInstances(&ec2.StopInstancesInput{
				InstanceIds:[]*string{
					descOutput.Reservations[0].Instances[0].InstanceId,
				},
			})
		}
		if strings.Split(value,":")[1]=="再起動"{
			_,err = ec.StartInstances(&ec2.StartInstancesInput{
				InstanceIds:[]*string{
					descOutput.Reservations[0].Instances[0].InstanceId,
				},
			})
		}
		if err != nil{
			return  response, errors.New("fail to operate instance")
		}

		log.Infof("originalMessage:%v",message.OriginalMessage)
		return makeResponse(&response, message.OriginalMessage, title, "")
	case actionCancel:
		title := fmt.Sprintf(":x: @%s はサーバーの状態変更を辞めたようだ", message.User.Name)
		log.Infof("originalMessage:%v",message.OriginalMessage)
		return makeResponse(&response, message.OriginalMessage, title, "")
	default:
		log.Printf("[ERROR] ]Invalid action was submitted: %s", action.Name)
		response.StatusCode =http.StatusInternalServerError
		return response,errors.New("Invalid action was submitted")
	}
}

//ボタンを空にして、前回のactionの結果を詰める
func makeResponse(response *events.APIGatewayProxyResponse,original slack.Message, title, value string) (events.APIGatewayProxyResponse,error){
	if original.Attachments == nil{
		original.Attachments = []slack.Attachment{slack.Attachment{}}
	}
	//テキストは上書きしない
	//original.Text =""
	original.Attachments[0].Actions = []slack.AttachmentAction{} // empty buttons
	original.Attachments[0].Fields = []slack.AttachmentField{
		{
			Title: title,
			Value: value,
			Short: false,
		},
	}
	resJson, err := json.Marshal(&original)
	if err != nil{
		return  *response, errors.New("fail to unmarshal originalMessage")
	}
	response.Body = string(resJson)
	response.Headers = make(map[string]string)
	response.Headers["Content-type"]= "application/json"
	response.StatusCode =http.StatusOK
	return  *response,nil
}

