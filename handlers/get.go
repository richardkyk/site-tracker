package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"site-tracker/clients"
	"strconv"

	"github.com/aws/aws-lambda-go/events"
	"github.com/gocolly/colly/v2"
)

type GetRequestBody struct {
	ID string `json:"id" validate:"required"`
}

func HandleGet(ctx context.Context, request events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	params := GetRequestBody{
		ID: request.QueryStringParameters["id"],
	}
	if err := Validate.Struct(params); err != nil {
		return events.APIGatewayV2HTTPResponse{
			StatusCode: 400,
			Body:       "validation failed: " + err.Error(),
			Headers:    map[string]string{"Content-Type": "plain/text"},
		}, nil
	}

	site, err := clients.GetItem(ctx, params.ID)
	if err != nil {
		return events.APIGatewayV2HTTPResponse{
			StatusCode: 500,
			Body:       err.Error(),
			Headers:    map[string]string{"Content-Type": "plain/text"},
		}, nil
	}

	c := colly.NewCollector()

	extractedValue := ""

	c.OnHTML(site.Selector, func(e *colly.HTMLElement) {
		text := e.Text
		valueRegex := regexp.MustCompile(site.Regex)
		match := valueRegex.FindStringSubmatch(text)
		extractedValue = "N/A"

		if len(match) > 1 {
			extractedValue = match[1]
		}

		if extractedValue != "N/A" && extractedValue != site.Expected {
			extractedValue = fmt.Sprintf(`Value changed from %s to %s`, strconv.Quote(site.Expected), strconv.Quote(extractedValue))
			payload := map[string]string{
				"email":   site.Email,
				"url":     site.URL,
				"message": extractedValue,
			}
			payloadBytes, err := json.Marshal(payload)
			if err != nil {
				log.Printf("failed to marshal payload, %v", err)
				return
			}

			if err := clients.SendMessage(ctx, string(payloadBytes)); err != nil {
				log.Printf("failed to send message to SQS queue, %v", err)
			}
			clients.UpdateItemShouldCheck(ctx, site.ID, false)

		}
	})

	c.OnRequest(func(r *colly.Request) {
		log.Println("visiting", r.URL.String())
	})

	c.OnError(func(r *colly.Response, err error) {
		log.Printf("request failed: %s %d %v", r.Request.URL, r.StatusCode, err)
		extractedValue = fmt.Sprintf("Error: %s", err.Error())
	})

	c.Visit(site.URL)

	return events.APIGatewayV2HTTPResponse{
		StatusCode:      200,
		Body:            extractedValue,
		IsBase64Encoded: false,
		Headers:         map[string]string{"Content-Type": "text/plain; charset=utf-8"},
	}, nil

}
