package status

type ScheduleHandlerStatus string

const (
	PathParameterError ScheduleHandlerStatus = "Failed to translate path parameter to uint64"
)
