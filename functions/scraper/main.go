package main

import (
	"context"
	"fmt"
	"site-tracker/functions/scraper/handlers"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

func router(ctx context.Context, request events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	method := request.RequestContext.HTTP.Method
	switch method {
	case "GET":
		return handlers.HandleGet(ctx, request)
	case "POST":
		return handlers.HandlePost(ctx, request)
	case "DELETE":
		return handlers.HandleDelete(ctx, request)
	case "OPTIONS":
		return handlers.HandleOptions(ctx, request)
	default:
		return events.APIGatewayV2HTTPResponse{
			StatusCode: 400,
			Body:       fmt.Sprintf("Invalid method: %s", method),
			Headers:    map[string]string{"Content-Type": "plain/text"},
		}, nil
	}
}

func main() {
	lambda.Start(router)
}
