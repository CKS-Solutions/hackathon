package notification

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/cks-solutions/hackathon/ms-video/internal/core/ports"
)

type NotificationServiceImpl struct {
	httpClient *http.Client
	baseURL    string
}

func NewNotificationService() ports.NotificationService {
	baseURL := os.Getenv("MS_NOTIFY_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8082"
	}

	return &NotificationServiceImpl{
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		baseURL: baseURL,
	}
}

type NotificationRequest struct {
	Subject string   `json:"subject"`
	To      []string `json:"to"`
	Html    string   `json:"html"`
}

func (n *NotificationServiceImpl) SendVideoProcessedNotification(ctx context.Context, email, videoID, originalName string) error {
	subject := "Video Processing Completed"
	html := fmt.Sprintf(`
		<html>
		<body>
			<h2>Video Processing Completed</h2>
			<p>Your video has been successfully processed!</p>
			<p><strong>Video ID:</strong> %s</p>
			<p><strong>Original Name:</strong> %s</p>
			<p>You can now download your processed video from the platform.</p>
			<br>
			<p>Thank you for using our service!</p>
		</body>
		</html>
	`, videoID, originalName)

	return n.sendNotification(ctx, email, subject, html)
}

func (n *NotificationServiceImpl) SendVideoFailedNotification(ctx context.Context, email, videoID, originalName, errorMessage string) error {
	subject := "Video Processing Failed"
	html := fmt.Sprintf(`
		<html>
		<body>
			<h2>Video Processing Failed</h2>
			<p>Unfortunately, your video processing has failed.</p>
			<p><strong>Video ID:</strong> %s</p>
			<p><strong>Original Name:</strong> %s</p>
			<p><strong>Error:</strong> %s</p>
			<p>Please try uploading your video again or contact our support team.</p>
			<br>
			<p>We apologize for the inconvenience.</p>
		</body>
		</html>
	`, videoID, originalName, errorMessage)

	return n.sendNotification(ctx, email, subject, html)
}

func (n *NotificationServiceImpl) sendNotification(ctx context.Context, email, subject, html string) error {
	notificationReq := NotificationRequest{
		Subject: subject,
		To:      []string{email},
		Html:    html,
	}

	jsonData, err := json.Marshal(notificationReq)
	if err != nil {
		return fmt.Errorf("failed to marshal notification request: %w", err)
	}

	url := fmt.Sprintf("%s/notify/notification", n.baseURL)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create notification request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := n.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send notification: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("notification service returned non-success status: %d", resp.StatusCode)
	}

	return nil
}
