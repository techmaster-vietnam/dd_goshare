package handlers

type ErrorItem struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}
type ErrorResponse struct {
	Errors ErrorItem `json:"errors"`
}

// SuccessResponse is the standard success payload for RESTful responses
// - Code mirrors the HTTP status code (e.g., 200, 201)
// - Message is optional but recommended for clarity
// - Data holds the response body (omit when not needed)
type SuccessResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}
