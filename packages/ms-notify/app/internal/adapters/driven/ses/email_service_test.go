package ses

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/ses"
	"github.com/cks-solutions/hackathon/ms-notify/internal/core/entities"
)

type fakeSESClient struct {
	sendEmailErr error
}

func (f *fakeSESClient) SendEmail(ctx context.Context, params *ses.SendEmailInput, optFns ...func(*ses.Options)) (*ses.SendEmailOutput, error) {
	if f.sendEmailErr != nil {
		return nil, f.sendEmailErr
	}
	return &ses.SendEmailOutput{}, nil
}

func TestEmailServiceImpl_Send(t *testing.T) {
	ctx := context.Background()
	notification := entities.Notification{
		Id: "1", Subject: "sub", From: "from@x.com", To: []string{"to@x.com"}, Html: "<p>body</p>",
	}

	t.Run("success", func(t *testing.T) {
		svc := NewEmailService(&fakeSESClient{}).(*EmailServiceImpl)
		err := svc.Send(ctx, notification)
		if err != nil {
			t.Errorf("Send: %v", err)
		}
	})

	t.Run("SendEmail error", func(t *testing.T) {
		svc := NewEmailService(&fakeSESClient{sendEmailErr: errors.New("ses error")}).(*EmailServiceImpl)
		err := svc.Send(ctx, notification)
		if err == nil {
			t.Error("expected error")
		}
	})
}
