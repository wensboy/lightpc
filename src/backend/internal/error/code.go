package error

const (
	// Root error codes.
	ErrCodeFatal        = "10010000"
	ErrCodeUnknown      = "10010001"
	ErrCodeNotExist     = "10010002"
	ErrCodeOpDenied     = "10010003"
	ErrCodeOpFailed     = "10010004"
	ErrCodeInvalidParam = "10010005"

	// Engine error codes.
	ErrCodeEngineInvalidPath = "10020000"
)
