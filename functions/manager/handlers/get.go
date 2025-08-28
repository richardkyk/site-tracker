package handlers

import (
	"context"
	"fmt"
	"site-tracker/internal/models"
	"site-tracker/internal/scraper"
	"strconv"

	"github.com/aws/aws-lambda-go/events"
	"github.com/google/uuid"
)

type GetRequestBody struct {
	URL      string `json:"url" validate:"required"`
	Selector string `json:"selector" validate:"required"`
	Regex    string `json:"regex" validate:"required"`
	Expected string `json:"expected" validate:"required"`
}

func HandleGet(ctx context.Context, request events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	params := GetRequestBody{
		URL:      request.QueryStringParameters["url"],
		Selector: request.QueryStringParameters["selector"],
		Regex:    request.QueryStringParameters["regex"],
		Expected: request.QueryStringParameters["expected"],
	}
	if err := Validate.Struct(params); err != nil {
		return events.APIGatewayV2HTTPResponse{
			StatusCode: 400,
			Body:       "validation failed: " + err.Error(),
			Headers:    map[string]string{"Content-Type": "text/plain"},
		}, nil
	}

	site := models.Site{
		ID:       uuid.New().String(),
		URL:      params.URL,
		Selector: params.Selector,
		Regex:    params.Regex,
		Expected: params.Expected,
	}

	extractedValue, err := scraper.Scrape(ctx, site)
	if err != nil {
		return events.APIGatewayV2HTTPResponse{
			StatusCode: 500,
			Body:       "failed to scrape site: " + err.Error(),
			Headers:    map[string]string{"Content-Type": "text/plain"},
		}, nil
	}

	if extractedValue != params.Expected {
		extractedValue = fmt.Sprintf(`Value changed from %s to %s`, strconv.Quote(params.Expected), strconv.Quote(extractedValue))
	}

	return events.APIGatewayV2HTTPResponse{
		StatusCode:      200,
		Body:            extractedValue,
		IsBase64Encoded: false,
		Headers:         map[string]string{"Content-Type": "text/plain; charset=utf-8"},
	}, nil

}
