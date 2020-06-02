package status

type HandlerStatus string

const (
	EmailInvalidError     HandlerStatus = "Email is invalid"
	GenericServerError    HandlerStatus = "Something went wrong"
	MalformedRequestError HandlerStatus = "Malformed request"
	PathParameterError    HandlerStatus = "Failed to translate path parameter to uint64"

	MetadataNotFoundError HandlerStatus = "Metadata not found"

	ScheduleDeleteSuccess HandlerStatus = "Schedule delete is successful"

	ScheduleCronFormatInvalidError    HandlerStatus = "Schedule cron format is invalid"
	ScheduleDuplicateJobNameArgsError HandlerStatus = "Schedule job name and args duplicate is found"
	ScheduleIDInvalidError            HandlerStatus = "Schedule ID is invalid"
	ScheduleGroupMissingError         HandlerStatus = "Schedule group is missing"
	ScheduleListNotFoundError         HandlerStatus = "Schedule list is not found"
	ScheduleTagMissingError           HandlerStatus = "Schedule tag(s) are missing"
)
