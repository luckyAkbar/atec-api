package common

import (
	stdhttp "github.com/sweet-go/stdlib/http"
)

// Error satisfy default Error interface
type Error struct {
	Message string
	Cause   error
	Code    int
	Type    error
}

// Error error message
func (e Error) Error() string {
	return e.Message
}

// GenerateStdlibHTTPResponse added functionality to self returning appropriate error message, setting accurate http status codes, and error code
func (e *Error) GenerateStdlibHTTPResponse(data any) *stdhttp.StandardResponse {
	return &stdhttp.StandardResponse{
		Success: false,
		Message: e.Message,
		Status:  e.Code,
		Error:   e.Type.Error(),
		Data:    data,
	}
}
