package entities

import (
	"testing"

	"github.com/cks-solutions/hackathon/ms-notify/internal/adapters/driver/dto"
)

func TestFromInput(t *testing.T) {
	input := dto.NotificationInput{
		Subject: "Test Subject",
		To:      []string{"a@b.com", "b@c.com"},
		Html:    "<p>body</p>",
	}
	n := FromInput(input)
	if n.Id == "" {
		t.Error("expected non-empty Id")
	}
	if n.Subject != input.Subject {
		t.Errorf("Subject = %q, want %q", n.Subject, input.Subject)
	}
	if n.From != EMAIL_SENDER {
		t.Errorf("From = %q, want %q", n.From, EMAIL_SENDER)
	}
	if len(n.To) != 2 || n.To[0] != "a@b.com" || n.To[1] != "b@c.com" {
		t.Errorf("To = %v", n.To)
	}
	if n.Html != input.Html {
		t.Errorf("Html = %q, want %q", n.Html, input.Html)
	}
}

func TestNotification_ToOutput(t *testing.T) {
	n := Notification{
		Id:      "id-1",
		Subject: "Sub",
		From:    "from@x.com",
		To:      []string{"to@y.com"},
		Html:    "html",
	}
	out := n.ToOutput()
	if out.Id != n.Id {
		t.Errorf("ToOutput().Id = %q, want %q", out.Id, n.Id)
	}
	if out.Subject != n.Subject {
		t.Errorf("ToOutput().Subject = %q, want %q", out.Subject, n.Subject)
	}
	if out.From != n.From {
		t.Errorf("ToOutput().From = %q, want %q", out.From, n.From)
	}
	if len(out.To) != 1 || out.To[0] != "to@y.com" {
		t.Errorf("ToOutput().To = %v", out.To)
	}
}

func TestNotification_ToDB(t *testing.T) {
	n := Notification{
		Id:      "id-2",
		Subject: "Sub2",
		From:    "f@x.com",
		To:      []string{"t@z.com"},
		Html:    "<b>bold</b>",
	}

	db := n.ToDB(NotificationSuccess)
	if db.Id != n.Id || db.Subject != n.Subject || db.From != n.From || db.Html != n.Html {
		t.Error("ToDB: id/subject/from/html should match Notification")
	}
	if db.Status != NotificationSuccess {
		t.Errorf("ToDB().Status = %q, want SUCCESS", db.Status)
	}

	dbFail := n.ToDB(NotificationFailure)
	if dbFail.Status != NotificationFailure {
		t.Errorf("ToDB(FAILURE).Status = %q", dbFail.Status)
	}
}
