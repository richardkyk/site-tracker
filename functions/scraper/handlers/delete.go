package handlers

import (
	"context"
	"encoding/json"
	"site-tracker/functions/scraper/clients"

	"github.com/aws/aws-lambda-go/events"
)

type DeleteRequestBody struct {
	Id string `json:"id" validate:"required"`
}

func HandleDelete(ctx context.Context, request events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	var body DeleteRequestBody
	if err := json.Unmarshal([]byte(request.Body), &body); err != nil {
		return events.APIGatewayV2HTTPResponse{
			StatusCode: 400,
			Body:       "invalid JSON: " + err.Error(),
			Headers:    map[string]string{"Content-Type": "plain/text"},
		}, nil
	}

	// Validate struct
	if err := Validate.Struct(body); err != nil {
		return events.APIGatewayV2HTTPResponse{
			StatusCode: 400,
			Body:       "validation failed: " + err.Error(),
			Headers:    map[string]string{"Content-Type": "plain/text"},
		}, nil
	}

	if err := clients.DeleteItem(ctx, body.Id); err != nil {
		return events.APIGatewayV2HTTPResponse{
			StatusCode: 500,
			Body:       err.Error(),
			Headers:    map[string]string{"Content-Type": "plain/text"},
		}, nil
	}

	return events.APIGatewayV2HTTPResponse{
		StatusCode: 201,
		Body:       "deleted: " + body.Id,
		Headers:    map[string]string{"Content-Type": "plain/text"},
	}, nil
}
