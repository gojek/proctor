package utility

const ClientError = "malformed request"
const ServerError = "Something went wrong"

const JobSubmissionSuccess = "success"
const JobSubmissionClientError = "client_error"
const JobSubmissionServerError = "server_error"

const JobNameContextKey = "job_name"
const JobArgsContextKey = "job_args"
const ImageNameContextKey = "image_name"
const JobSubmittedForExecutionContextKey = "job_submitted_for_execution"
const JobSubmissionStatusContextKey = "job_sumission_status"

func MergeMaps(mapOne, mapTwo map[string]string) map[string]string {
	result := make(map[string]string)

	for k, v := range mapOne {
		result[k] = v
	}
	for k, v := range mapTwo {
		result[k] = v
	}
	return result
}
