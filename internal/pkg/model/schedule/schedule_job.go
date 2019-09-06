package schedule

type ScheduledJob struct {
	ID                 uint64            `json:"id"`
	Name               string            `json:"name"`
	Args               map[string]string `json:"args"`
	NotificationEmails string            `json:"notification_emails"`
	Cron               string            `json:"cron"`
	Tags               string            `json:"tags"`
	Group              string            `json:"group_name"`
}
