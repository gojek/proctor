package schedule

import (
	"proctor/proctord/storage/postgres"
	"proctor/proctord/utility"
)

type ScheduledJob struct {
	ID                 string            `json:"id"`
	Name               string            `json:"name"`
	Args               map[string]string `json:"args"`
	NotificationEmails string            `json:"notification_emails"`
	Time               string            `json:"time"`
	Tags               string            `json:"tags"`
	Group              string            `json:"group_name"`
}

func FromStoreToHandler(scheduledJobsStoreFormat []postgres.JobsSchedule) ([]ScheduledJob, error) {
	var scheduledJobs []ScheduledJob
	for _, scheduledJobStoreFormat := range scheduledJobsStoreFormat {
		scheduledJob, err := GetScheduledJob(scheduledJobStoreFormat)
		if err != nil {
			return nil, err
		}
		scheduledJobs = append(scheduledJobs, scheduledJob)
	}
	return scheduledJobs, nil
}

func GetScheduledJob(scheduledJobStoreFormat postgres.JobsSchedule) (ScheduledJob, error) {
	args, err := utility.DeserializeMap(scheduledJobStoreFormat.Args)
	if err != nil {
		return ScheduledJob{}, err
	}
	scheduledJob := ScheduledJob{
		ID:                 scheduledJobStoreFormat.ID,
		Name:               scheduledJobStoreFormat.Name,
		Args:               args,
		Tags:               scheduledJobStoreFormat.Tags,
		Time:               scheduledJobStoreFormat.Time,
		Group:              scheduledJobStoreFormat.Group,
		NotificationEmails: scheduledJobStoreFormat.NotificationEmails,
	}
	return scheduledJob, nil

}
