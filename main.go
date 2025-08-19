package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"regexp"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/gocolly/colly/v2"
	"github.com/google/uuid"
)

var (
	dynamoDBClient *dynamodb.Client
	sqsClient      *sqs.Client
)

func init() {
	// Initialize the DynamoDB client outside of the handler, during the init phase
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}
	dynamoDBClient = dynamodb.NewFromConfig(cfg)
	sqsClient = sqs.NewFromConfig(cfg)
}

func handler(ctx context.Context, event events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	response := ""
	var err error
	method := event.RequestContext.HTTP.Method
	fmt.Println(method)
	switch method {
	case "GET":
		response, err = get(ctx, event)
	case "POST":
		response, err = post(ctx, event)
	default:
		err = fmt.Errorf("Invalid method")
	}

	if err != nil {
		return events.APIGatewayV2HTTPResponse{
			StatusCode: 400,
			Body:       err.Error(),
		}, nil
	}

	return events.APIGatewayV2HTTPResponse{
		StatusCode:      200,
		Headers:         map[string]string{"Content-Type": "text/plain; charset=utf-8"},
		Body:            response,
		IsBase64Encoded: false,
	}, nil

}

func main() {
	lambda.Start(handler)
}

type Event struct {
	ID string `json:"id"`
}

func get(ctx context.Context, event events.APIGatewayV2HTTPRequest) (string, error) {
	id := event.QueryStringParameters["id"]
	if id == "" {
		return "", fmt.Errorf("id is required")
	}

	// Get item from DynamoDB
	out, err := dynamoDBClient.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(os.Getenv("DYNAMODB_TABLE")),
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: id},
		},
	})
	if err != nil {
		return "", err
	}

	// Convert DynamoDB item to simple map[string]string
	item := make(map[string]string)
	for k, v := range out.Item {
		if s, ok := v.(*types.AttributeValueMemberS); ok {
			item[k] = s.Value
		}
	}

	c := colly.NewCollector()

	url := item["url"]
	selector := item["selector"]
	regex := item["regex"]
	expected := item["expected"]
	email := item["email"]

	extractedValue := ""
	valueRegex := regexp.MustCompile(regex)

	c.OnHTML(selector, func(e *colly.HTMLElement) {
		text := e.Text
		match := valueRegex.FindStringSubmatch(text)
		extractedValue = "N/A"
		if len(match) > 1 {
			extractedValue = match[1]
		}
		if extractedValue != "N/A" && extractedValue != expected {
			payload := map[string]string{
				"email":   email,
				"message": fmt.Sprintf("Value changed from %s to %s", expected, extractedValue),
			}
			payloadBytes, err := json.Marshal(payload)
			if err != nil {
				log.Printf("failed to marshal payload, %v", err)
				return
			}
			_, err = sqsClient.SendMessage(ctx, &sqs.SendMessageInput{
				QueueUrl:    aws.String(os.Getenv("SQS_URL")),
				MessageBody: aws.String(string(payloadBytes)),
			})
			if err != nil {
				log.Printf("failed to send message to SQS queue, %v", err)
			}
		}
	})

	c.OnRequest(func(r *colly.Request) {
		log.Println("visiting", r.URL.String())
	})

	c.OnError(func(r *colly.Response, err error) {
		log.Printf("Request failed: %s %d %v", r.Request.URL, r.StatusCode, err)
		extractedValue = fmt.Sprintf("Error: %s", err.Error())
	})

	c.Visit(url)
	log.Printf("value: %s", extractedValue)
	return extractedValue, nil

}

type RequestBody struct {
	URL      string `json:"url"`
	Selector string `json:"selector"`
	Regex    string `json:"regex"`
	Expected string `json:"expected"`
	Email    string `json:"email"`
}

func post(ctx context.Context, event events.APIGatewayV2HTTPRequest) (string, error) {
	// Parse POST body
	var body RequestBody
	err := json.Unmarshal([]byte(event.Body), &body)
	if err != nil {
		return "", fmt.Errorf("failed to parse body: %w", err)
	}
	if body.URL == "" {
		return "", fmt.Errorf("url is required")
	}

	if body.Selector == "" {
		return "", fmt.Errorf("selector is required")
	}

	if body.Regex == "" {
		return "", fmt.Errorf("regex is required")
	}

	// Generate a UUID for the item
	id := uuid.New().String()

	tableName := os.Getenv("DYNAMODB_TABLE")
	// Put item into DynamoDB
	_, err = dynamoDBClient.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(tableName),
		Item: map[string]types.AttributeValue{
			"id":       &types.AttributeValueMemberS{Value: id},
			"url":      &types.AttributeValueMemberS{Value: body.URL},
			"selector": &types.AttributeValueMemberS{Value: body.Selector},
			"regex":    &types.AttributeValueMemberS{Value: body.Regex},
			"expected": &types.AttributeValueMemberS{Value: body.Expected},
			"email":    &types.AttributeValueMemberS{Value: body.Email},
		},
	})
	if err != nil {
		return "", fmt.Errorf("failed to put item: %w", err)
	}

	// Return the generated ID
	return id, nil
}
