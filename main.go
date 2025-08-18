package main

import (
	"context"
	"fmt"
	"log"
	"regexp"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/gocolly/colly/v2"
)

type QueryParams struct {
	Url      string `json:"url"`
	Selector string `json:"selector"`
	Regex    string `json:"regex"`
}

func handler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	params, err := parseQueryParams(request.QueryStringParameters)

	if err != nil {
		log.Printf("Failed to parse input: %v", err)
		return events.APIGatewayProxyResponse{
			StatusCode: 400,
			Body:       err.Error(),
		}, nil
	}

	c := colly.NewCollector()

	price := ""
	priceRegex := regexp.MustCompile(params.Regex)

	c.OnHTML(params.Selector, func(e *colly.HTMLElement) {
		text := e.Text
		match := priceRegex.FindStringSubmatch(text)
		price = "N/A"
		if len(match) > 1 {
			price = match[1]
		}
	})

	c.OnRequest(func(r *colly.Request) {
		log.Println("visiting", r.URL.String())
	})

	c.OnError(func(r *colly.Response, err error) {
		log.Printf("Request failed: %s %d %v", r.Request.URL, r.StatusCode, err)
		price = fmt.Sprintf("Error: %s", err.Error())
	})

	c.Visit(params.Url)
	log.Printf("price: %s", price)

	return events.APIGatewayProxyResponse{
		StatusCode:      200,
		Headers:         map[string]string{"Content-Type": "text/plain; charset=utf-8"},
		Body:            price,
		IsBase64Encoded: false,
	}, nil
}

func main() {
	lambda.Start(handler)
}

func parseQueryParams(q map[string]string) (QueryParams, error) {
	url := q["url"]
	if url == "" {
		return QueryParams{}, fmt.Errorf("missing param: url")
	}

	selector := q["selector"]
	if selector == "" {
		return QueryParams{}, fmt.Errorf("missing param: selector")
	}

	regex := q["regex"]
	if regex == "" {
		return QueryParams{}, fmt.Errorf("missing param: regex")
	}

	return QueryParams{
		Url:      url,
		Selector: selector,
		Regex:    regex,
	}, nil
}
