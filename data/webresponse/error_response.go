package webresponse

type ErrorResponse struct {
	Description string `json:"description"`
	Data        any    `json:"data,omitempty"`
}
