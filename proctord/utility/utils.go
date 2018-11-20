package utility

const ClientError = "malformed request"
const ServerError = "Something went wrong"

const UnauthorizedErrorMissingConfig = "EMAIL_ID or ACCESS_TOKEN is not present in proctor config file."
const UnauthorizedErrorInvalidConfig = "Please check the EMAIL_ID and ACCESS_TOKEN validity in proctor config file."
const GenericListCmdError = "Error fetching list of procs. Please check configuration and network connectivity"
const GenericProcCmdError = "Error executing proc. Please check configuration and network connectivity"
const GenericDescribeCmdError = "Error fetching description of proc. Please check configuration and network connectivity"

const UnauthorizedErrorHeader = "Unauthorized Access!!!"
const GenericTimeoutErrorHeader = "Connection Timeout!!!"
const GenericNetworkErrorHeader = "Network Error!!!"
const GenericResponseErrorHeader = "Server Error!!!"

const ConfigProctorHostMissingError = "Config Error!!!\nMandatory config PROCTOR_HOST is missing in Proctor Config file."
const GenericTimeoutErrorBody = "Please check your Internet/VPN connection for connectivity to ProctorD."

const JobSubmissionSuccess = "success"
const JobSubmissionClientError = "client_error"
const JobSubmissionServerError = "server_error"

const JobSucceeded = "SUCCEEDED"
const JobFailed = "FAILED"
const JobWaiting = "WAITING"

const JobNameContextKey = "job_name"
const UserEmailContextKey = "user_email"
const JobArgsContextKey = "job_args"
const ImageNameContextKey = "image_name"
const JobNameSubmittedForExecutionContextKey = "job_name_submitted_for_execution"
const JobSubmissionStatusContextKey = "job_sumission_status"

const UserEmailHeaderKey = "Email-Id"
const AccessTokenHeaderKey = "Access-Token"
const ClientVersion  =  "Client-Version"

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
