package ses

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/ses"
	"github.com/aws/aws-sdk-go-v2/service/ses/types"
	"github.com/cks-solutions/hackathon/ms-notify/internal/core/entities"
	"github.com/cks-solutions/hackathon/ms-notify/internal/core/ports"
	"github.com/cks-solutions/hackathon/ms-notify/internal/infra/aws"
)

type EmailServiceImpl struct {
	client aws.SESClient
}

func NewEmailService(client aws.SESClient) ports.EmailService {
	return &EmailServiceImpl{client: client}
}

func (s *EmailServiceImpl) Send(ctx context.Context, notification entities.Notification) error {
	_, err := s.client.SendEmail(ctx, &ses.SendEmailInput{
		Source: &notification.From,
		Destination: &types.Destination{
			ToAddresses: notification.To,
		},
		Message: &types.Message{
			Body: &types.Body{
				Html: &types.Content{
					Data: &notification.Html,
				},
			},
			Subject: &types.Content{
				Data: &notification.Subject,
			},
		},
	})
	if err != nil {
		return err
	}

	return nil
}
