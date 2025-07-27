package response

type InternalServerError struct {
	Message   string `json:"message"`
	RequestID string `json:"request_id"`
}

func (e InternalServerError) Error() string {
	return e.Message
}
