package constants

const (
	// SuccessCode 通用响应
	SuccessCode = "S00000"

	// FailCode 通用错误
	FailCode = "E00000"

	// ParamError 参数类错误。统一用E1开头
	ParamError        = "E10001"
	MissingFieldError = "E10002"

	// UserNotFound 业务类错误，统一用E2开头
	UserNotFound     = "E20001"
	PermissionDenied = "E20002"

	AccountError = "E30001"

	ContractError = "E40001"

	// InternalServerError 系统内部错误，非代码逻辑错误。
	InternalServerError = "E90000"
)
