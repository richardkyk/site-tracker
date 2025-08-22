package sqs

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

type SQSClientWrapper struct {
	Client   *sqs.Client
	QueueURL string
}

// NewSQSClient returns a new SQS client for the given queue URL
func NewSQSClient(queueURL string) (*SQSClientWrapper, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return nil, fmt.Errorf("unable to load SDK config: %w", err)
	}

	return &SQSClientWrapper{
		Client:   sqs.NewFromConfig(cfg),
		QueueURL: queueURL,
	}, nil
}

func (s *SQSClientWrapper) SendMessage(ctx context.Context, message string) error {
	_, err := s.Client.SendMessage(ctx, &sqs.SendMessageInput{
		QueueUrl:    &s.QueueURL,
		MessageBody: &message,
	})
	if err != nil {
		return fmt.Errorf("failed to send message to SQS queue, %w", err)
	}
	return nil
}
