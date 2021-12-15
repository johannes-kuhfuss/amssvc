package dto

type NewJobRequest struct {
	Name   string `json:"name"`
	SrcUrl string `json:"src_url"`
}
