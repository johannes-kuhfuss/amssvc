package dto

type JobStatusUpdateRequest struct {
	Status string `json:"status"`
	ErrMsg string `json:"err_msg"`
}
