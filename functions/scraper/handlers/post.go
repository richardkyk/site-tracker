package handlers

import (
	"context"
	"encoding/json"
	"site-tracker/functions/scraper/clients"
	"site-tracker/functions/scraper/models"

	"github.com/aws/aws-lambda-go/events"
	"github.com/google/uuid"
)

type PostRequestBody struct {
	URL      string `json:"url" validate:"required,url"`
	Selector string `json:"selector" validate:"required"`
	Regex    string `json:"regex" validate:"required"`
	Expected string `json:"expected"`
	Email    string `json:"email" validate:"required,email"`
}

func HandlePost(ctx context.Context, request events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	// Parse POST body
	var body PostRequestBody
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

	// Generate a UUID for the item
	id := uuid.New().String()
	site := models.Site{
		ID:          id,
		URL:         body.URL,
		Selector:    body.Selector,
		Regex:       body.Regex,
		Expected:    body.Expected,
		Email:       body.Email,
		ShouldCheck: true,
	}
	if err := clients.PutItem(ctx, site); err != nil {
		return events.APIGatewayV2HTTPResponse{
			StatusCode: 500,
			Body:       err.Error(),
			Headers:    map[string]string{"Content-Type": "plain/text"},
		}, nil
	}

	return events.APIGatewayV2HTTPResponse{
		StatusCode: 201,
		Body:       id,
		Headers:    map[string]string{"Content-Type": "plain/text"},
	}, nil
}
