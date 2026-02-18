package entities

import (
	"github.com/cks-solutions/hackathon/ms-notify/internal/adapters/driver/dto"
	"github.com/google/uuid"
)

type Notification struct {
	Id      string   `json:"id"`
	Subject string   `json:"subject"`
	From    string   `json:"from"`
	To      []string `json:"to"`
	Html    string   `json:"html"`
}

type NotificationStatus string

const (
	NotificationSuccess NotificationStatus = "SUCCESS"
	NotificationFailure NotificationStatus = "FAILURE"
)

const EMAIL_SENDER = "cks.hackathon.noreply@gmail.com"

type NotificationDB struct {
	Id      string             `dynamodbav:"id"`
	Subject string             `dynamodbav:"subject"`
	From    string             `dynamodbav:"from"`
	To      []string           `dynamodbav:"to"`
	Html    string             `dynamodbav:"html"`
	Status  NotificationStatus `dynamodbav:"status"`
}

func FromInput(input dto.NotificationInput) Notification {
	return Notification{
		Id:      uuid.NewString(),
		Subject: input.Subject,
		From:    EMAIL_SENDER,
		To:      input.To,
		Html:    input.Html,
	}
}

func (n *Notification) ToOutput() dto.NotificationOutput {
	return dto.NotificationOutput{
		Id:      n.Id,
		Subject: n.Subject,
		From:    n.From,
		To:      n.To,
	}
}

func (n *Notification) ToDB(status NotificationStatus) NotificationDB {
	return NotificationDB{
		Id:      n.Id,
		Subject: n.Subject,
		From:    n.From,
		To:      n.To,
		Html:    n.Html,
		Status:  status,
	}
}
