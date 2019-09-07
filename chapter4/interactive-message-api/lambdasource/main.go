package main

import (
	"context"
	"errors"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"log"
	"syscall"
)

func main() {
	lambda.Start(handleRequest)
}

func handleRequest(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	response := events.APIGatewayProxyResponse{}

	botOauthToken, found := syscall.Getenv("BOT_TOKEN")
	if !found {
		log.Print("BOT_TOKEN Not Found")
		return response,errors.New("BOT_TOKEN Not Found")
	}

	interactionHandler := NewInteractionHandler(botOauthToken)
	response,err := interactionHandler.LambdaHandle(request)
	if err != nil {
		return response, err
	}
	return response, nil
}



