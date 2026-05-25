package response

type Response struct {
	Success    bool         `json:"success"`
	Data       interface{}  `json:"data,omitempty"`
	Error      *ErrorDetail `json:"error,omitempty"`
	Idempotent bool         `json:"idempotent,omitempty"`
}

type ErrorDetail struct {
	Code    string       `json:"code"`
	Message string       `json:"message"`
	Details []FieldError `json:"details,omitempty"`
}

type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func Success(data interface{}) Response {
	return Response{Success: true, Data: data}
}

func SuccessIdempotent(data interface{}) Response {
	return Response{Success: true, Data: data, Idempotent: true}
}

func Error(code, message string) Response {
	return Response{Success: false, Error: &ErrorDetail{Code: code, Message: message}}
}

func ValidationError(message string, details []FieldError) Response {
	return Response{
		Success: false,
		Error:   &ErrorDetail{Code: "VALIDATION_ERROR", Message: message, Details: details},
	}
}
