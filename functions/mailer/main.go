package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"site-tracker/internal/clients/ses"
	"site-tracker/internal/models"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

func handler(ctx context.Context, sqsEvent events.SQSEvent) (events.SQSEventResponse, error) {
	response := events.SQSEventResponse{}

	sesClient, err := ses.NewClient(os.Getenv("SES_FROM_EMAIL"))
	if err != nil {
		return response, err
	}
	for _, record := range sqsEvent.Records {
		log.Printf("Message ID: %s, Body: %s", record.MessageId, record.Body)

		var email models.Email
		if err := json.Unmarshal([]byte(record.Body), &email); err != nil {
			log.Printf("failed to unmarshal SES message: %v", err)
			continue // skip bad message
		}

		body := fmt.Sprintf("Status: %s\n\n%s\n\n%s", email.Status, email.URL, email.Message)

		if err := sesClient.SendEmail(ctx, email.Email, "Site Tracker", body); err != nil {
			log.Printf("failed to send message to SQS queue, %v", err)
		}
	}

	return response, nil
}

func main() {
	lambda.Start(handler)
}
