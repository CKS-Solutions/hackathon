package dto

type NotificationInput struct {
	Subject string   `json:"subject"`
	To      []string `json:"to"`
	Html    string   `json:"html"`
}

type NotificationOutput struct {
	Id      string   `json:"id"`
	Subject string   `json:"subject"`
	From    string   `json:"from"`
	To      []string `json:"to"`
}
