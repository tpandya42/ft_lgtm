package api

type ExecuteRequest struct {
	Code     string `json:"code"`
	Language string `json:"language"`
}

type ExecuteResponse struct {
	Success bool   `json:"success"`
	Output  string `json:"output,omitempty"`
	Error   string `json:"error,omitempty"`
	Stage   string `json:"stage,omitempty"`
}
