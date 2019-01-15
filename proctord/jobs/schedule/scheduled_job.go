package schedule

type ScheduledJob struct {
	ID                 string            `json:"id"`
	Name               string            `json:"name"`
	Args               map[string]string `json:"args"`
	NotificationEmails string            `json:"notification_emails"`
	Time               string            `json:"time"`
	Tags               string            `json:"tags"`
}
