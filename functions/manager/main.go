package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"site-tracker/functions/manager/handlers"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

func router(ctx context.Context, request events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	method := request.RequestContext.HTTP.Method
	log.Printf("method: %s", method)

	switch method {
	case http.MethodGet:
		return handlers.HandleGet(ctx, request)
	case http.MethodPost:
		return handlers.HandlePost(ctx, request)
	case http.MethodDelete:
		return handlers.HandleDelete(ctx, request)
	case http.MethodOptions:
		return handlers.HandleOptions(ctx, request)
	default:
		return events.APIGatewayV2HTTPResponse{
			StatusCode: http.StatusBadRequest,
			Body:       fmt.Sprintf("Invalid method: %s", method),
			Headers:    map[string]string{"Content-Type": "text/plain"},
		}, nil
	}
}

func main() {
	lambda.Start(router)
}
