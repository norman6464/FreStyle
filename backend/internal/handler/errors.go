package handler

// errorResponse は全 handler 共通のエラー JSON 形式。
type errorResponse struct {
	Error string `json:"error"`
}
