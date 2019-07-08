package schedule

import (
	"proctor/internal/app/proctord/storage/postgres"
	modelSchedule "proctor/internal/pkg/model/schedule"
	"proctor/internal/pkg/utility"
)

func FromStoreToHandler(scheduledJobsStoreFormat []postgres.JobsSchedule) ([]modelSchedule.ScheduledJob, error) {
	var scheduledJobs []modelSchedule.ScheduledJob
	for _, scheduledJobStoreFormat := range scheduledJobsStoreFormat {
		scheduledJob, err := GetScheduledJob(scheduledJobStoreFormat)
		if err != nil {
			return nil, err
		}
		scheduledJobs = append(scheduledJobs, scheduledJob)
	}
	return scheduledJobs, nil
}

func GetScheduledJob(scheduledJobStoreFormat postgres.JobsSchedule) (modelSchedule.ScheduledJob, error) {
	args, err := utility.DeserializeMap(scheduledJobStoreFormat.Args)
	if err != nil {
		return modelSchedule.ScheduledJob{}, err
	}
	scheduledJob := modelSchedule.ScheduledJob{
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
