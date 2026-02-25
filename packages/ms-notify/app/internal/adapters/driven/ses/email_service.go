package ses

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/ses"
	"github.com/aws/aws-sdk-go-v2/service/ses/types"
	"github.com/cks-solutions/hackathon/ms-notify/internal/core/entities"
	"github.com/cks-solutions/hackathon/ms-notify/internal/core/ports"
)

// sesClientInterface allows testing without a real SES client. *infra/aws.SESClient satisfies it.
type sesClientInterface interface {
	SendEmail(ctx context.Context, params *ses.SendEmailInput, optFns ...func(*ses.Options)) (*ses.SendEmailOutput, error)
}

type EmailServiceImpl struct {
	client sesClientInterface
}

func NewEmailService(client sesClientInterface) ports.EmailService {
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
