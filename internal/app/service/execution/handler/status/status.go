package status

type ExecutionHandlerStatus string

const (
	MalformedRequest         ExecutionHandlerStatus = "Failed to parse request body from CLI"
	JobExecutionError        ExecutionHandlerStatus = "Failed to execute Job into Executor"
	PathParameterError       ExecutionHandlerStatus = "Failed to translate path parameter to uint64"
	WebSocketInitError       ExecutionHandlerStatus = "WebSocket Initializer Error"
	ExecutionContextNotFound ExecutionHandlerStatus = "Execution context not Found"
)
