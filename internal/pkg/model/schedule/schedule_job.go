package schedule

type ScheduledJob struct {
	ID                 uint64            `json:"id"`
	Name               string            `json:"jobName"`
	Args               map[string]string `json:"args"`
	NotificationEmails string            `json:"notificationEmails"`
	Cron               string            `json:"cron"`
	Tags               string            `json:"tags"`
	Group              string            `json:"group"`
	Enabled            bool              `json:"enabled"`
}
