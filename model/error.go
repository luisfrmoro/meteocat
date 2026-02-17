package model

// APIError represents an error returned by the METEOCAT API or encountered while performing a request.
// When no HTTP response was received, the Code field will be zero.
type APIError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (e *APIError) Error() string {
	return e.Message
}
