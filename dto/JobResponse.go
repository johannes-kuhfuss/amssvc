package dto

import (
	"time"
)

type JobResponse struct {
	Id         string    `json:"job_id"`
	Name       string    `json:"name"`
	CreatedAt  time.Time `json:"created_at"`
	CreatedBy  string    `json:"created_by"`
	ModifiedAt time.Time `json:"modified_at"`
	ModifiedBy string    `json:"modified_by"`
	SrcUrl     string    `json:"src_url"`
	Status     string    `json:"status"`
	ErrorMsg   string    `json:"error_msg"`
	TechInfo   string    `json:"tech_info"`
}
