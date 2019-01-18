package schedule

import (
	"github.com/gojektech/proctor/proctord/logger"
	"github.com/gojektech/proctor/proctord/storage/postgres"
	"github.com/gojektech/proctor/proctord/utility"
)

type ScheduledJob struct {
	ID                 string            `json:"id"`
	Name               string            `json:"name"`
	Args               map[string]string `json:"args"`
	NotificationEmails string            `json:"notification_emails"`
	Time               string            `json:"time"`
	Tags               string            `json:"tags"`
}

func FromStoreToHandler(scheduledJobsStoreFormat []postgres.JobsSchedule) []ScheduledJob {
	var scheduledJobs []ScheduledJob
	for _, scheduledJobStoreFormat := range scheduledJobsStoreFormat {
		args, err := utility.DeserializeMap(scheduledJobStoreFormat.Args)
		if err != nil {
			logger.Error("Error deserializing scheduled job args to map: ", err.Error())
		}
		scheduledJob := ScheduledJob{
			ID:                 scheduledJobStoreFormat.ID,
			Name:               scheduledJobStoreFormat.Name,
			Args:               args,
			Tags:               scheduledJobStoreFormat.Tags,
			Time:               scheduledJobStoreFormat.Time,
			NotificationEmails: scheduledJobStoreFormat.NotificationEmails,
		}
		scheduledJobs = append(scheduledJobs, scheduledJob)
	}
	return scheduledJobs
}
