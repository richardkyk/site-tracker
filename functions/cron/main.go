package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"site-tracker/internal/clients/dynamodb"
	"site-tracker/internal/clients/sqs"
	"site-tracker/internal/models"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
)

func handler(ctx context.Context) (events.APIGatewayV2HTTPResponse, error) {
	output, err := dynamodb.Scan(ctx)
	if err != nil {
		return events.APIGatewayV2HTTPResponse{
			StatusCode: 500,
			Body:       err.Error(),
			Headers:    map[string]string{"Content-Type": "text/plain"},
		}, nil
	}

	log.Printf("Found %d items", len(output.Items))
	successCount := 0

	if len(output.Items) == 0 {
		return events.APIGatewayV2HTTPResponse{
			StatusCode: 200,
			Body:       "No items found",
			Headers:    map[string]string{"Content-Type": "text/plain"},
		}, nil
	}

	sqsClient, err := sqs.NewClient(os.Getenv("SQS_TASK_URL"))
	if err != nil {
		log.Printf("failed to create SQS client, %v", err)
		return events.APIGatewayV2HTTPResponse{
			StatusCode: 500,
			Body:       err.Error(),
			Headers:    map[string]string{"Content-Type": "text/plain"},
		}, nil
	}

	for _, item := range output.Items {
		var site models.Site
		err := attributevalue.UnmarshalMap(item, &site)
		if err != nil {
			log.Printf("failed to unmarshal DynamoDB item: %v", err)
			continue
		}

		siteJSON, err := json.Marshal(site)
		if err != nil {
			log.Printf("failed to marshal item, %v", err)
			continue
		}
		err = sqsClient.SendMessage(ctx, string(siteJSON))
		if err == nil {
			successCount++
		}
	}

	return events.APIGatewayV2HTTPResponse{
		StatusCode:      200,
		Body:            fmt.Sprintf("Checking %d sites", successCount),
		IsBase64Encoded: false,
		Headers:         map[string]string{"Content-Type": "text/plain; charset=utf-8"},
	}, nil
}

func main() {
	lambda.Start(handler)
}
