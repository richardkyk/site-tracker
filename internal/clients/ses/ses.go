package ses

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ses"
	sesTypes "github.com/aws/aws-sdk-go-v2/service/ses/types"
)

type SESClientWrapper struct {
	Client *ses.Client
	From   string
}

// NewSESClient returns a new SES client
func NewClient(from string) (*SESClientWrapper, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return nil, fmt.Errorf("unable to load SDK config: %w", err)
	}

	return &SESClientWrapper{
		Client: ses.NewFromConfig(cfg),
		From:   from,
	}, nil
}

func (s *SESClientWrapper) SendEmail(ctx context.Context, to string, subject string, body string) error {
	_, err := s.Client.SendEmail(ctx, &ses.SendEmailInput{
		Source: &s.From,
		Destination: &sesTypes.Destination{
			ToAddresses: []string{to},
		},
		Message: &sesTypes.Message{
			Subject: &sesTypes.Content{
				Data: &subject,
			},
			Body: &sesTypes.Body{
				Text: &sesTypes.Content{
					Data: &body,
				},
			},
		},
	})
	if err != nil {
		return fmt.Errorf("failed to send email with SES: %w", err)
	}
	return nil
}
