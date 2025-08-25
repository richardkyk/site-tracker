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
	"site-tracker/internal/scraper"
	"strconv"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

func handler(ctx context.Context, sqsEvent events.SQSEvent) (events.SQSEventResponse, error) {
	response := events.SQSEventResponse{}

	sqsClient, err := sqs.NewClient(os.Getenv("SQS_NOTIFY_URL"))
	if err != nil {
		return response, err
	}

	for _, record := range sqsEvent.Records {
		log.Printf("Message ID: %s, Body: %s", record.MessageId, record.Body)

		var site models.Site
		if err := json.Unmarshal([]byte(record.Body), &site); err != nil {
			log.Printf("failed to unmarshal SQS message: %v", err)
			continue // skip bad message
		}

		message, status := "", ""
		extractedValue, err := scraper.Scrape(site)
		if err != nil {
			message = fmt.Sprintf("failed to scrape: %s", err.Error())
			status = "failed"
		} else if extractedValue != site.Expected {
			message = fmt.Sprintf(`value changed from %s to %s`, strconv.Quote(site.Expected), strconv.Quote(extractedValue))
			status = "changed"
		}

		if message != "" {
			payload := models.Email{
				Email:   site.Email,
				URL:     site.URL,
				Message: message,
				Status:  status,
			}
			payloadBytes, err := json.Marshal(payload)
			if err != nil {
				log.Printf("failed to marshal payload, %v", err)

				response.BatchItemFailures = append(response.BatchItemFailures, events.SQSBatchItemFailure{
					ItemIdentifier: record.MessageId,
				})
			}
			if err := sqsClient.SendMessage(ctx, string(payloadBytes)); err != nil {
				log.Printf("failed to send message to SQS queue, %v", err)
			}
			if err := dynamodb.UpdateItemShouldCheck(ctx, site.ID, false); err != nil {
				log.Printf("failed to update item, %v", err)
			}
		}
	}

	return response, nil
}

func main() {
	lambda.Start(handler)
}
