package domain

import (
	"fmt"
	"strings"
	"time"

	"github.com/johannes-kuhfuss/amssvc/dto"
	"github.com/johannes-kuhfuss/services_utils/api_error"
	"github.com/johannes-kuhfuss/services_utils/date"
	"github.com/segmentio/ksuid"
)

type JobStatus string

const (
	JobStatusCreated  JobStatus = "created"
	JobStatusQueued   JobStatus = "queued"
	JobStatusRunning  JobStatus = "running"
	JobStatusPaused   JobStatus = "paused"
	JobStatusFinished JobStatus = "finished"
	JobStatusFailed   JobStatus = "failed"
)

type Job struct {
	Id         ksuid.KSUID `db:"job_id"`
	Name       string      `db:"name"`
	CreatedAt  time.Time   `db:"created_at"`
	CreatedBy  string      `db:"created_by"`
	ModifiedAt time.Time   `db:"modified_at"`
	ModifiedBy string      `db:"modified_by"`
	SrcUrl     string      `db:"src_url"`
	Status     JobStatus   `db:"status"`
	ErrorMsg   string      `db:"error_msg"`
}

type JobStatusUpdate struct {
	newStatus JobStatus
	errMsg    string
}

type JobRepository interface {
	FindAll(string) (*[]Job, api_error.ApiErr)
	FindById(string) (*Job, api_error.ApiErr)
	Save(Job) api_error.ApiErr
	DeleteById(string) api_error.ApiErr
	GetNext() (*Job, api_error.ApiErr)
	SetStatus(string, JobStatusUpdate) api_error.ApiErr
}

func createJobName(name string) string {
	var jobName string
	if strings.TrimSpace(name) == "" {
		newDate, _ := date.GetNowLocalString("")
		jobName = fmt.Sprintf("new job @ %s", *newDate)
	} else {
		jobName = name
	}
	return jobName
}

func NewJob(name string, srcurl string) (*Job, api_error.ApiErr) {
	if strings.TrimSpace(srcurl) == "" {
		return nil, api_error.NewBadRequestError("Job must have a source URL")
	}

	return &Job{
		Id:         ksuid.New(),
		Name:       createJobName(name),
		CreatedAt:  date.GetNowUtc(),
		CreatedBy:  "",
		ModifiedAt: date.GetNowUtc(),
		ModifiedBy: "",
		SrcUrl:     srcurl,
		Status:     JobStatusCreated,
		ErrorMsg:   "",
	}, nil
}

func (job Job) ToDto() dto.JobResponse {
	return dto.JobResponse{
		Id:         job.Id.String(),
		Name:       job.Name,
		CreatedAt:  job.CreatedAt,
		CreatedBy:  job.CreatedBy,
		ModifiedAt: job.ModifiedAt,
		ModifiedBy: job.ModifiedBy,
		SrcUrl:     job.SrcUrl,
		Status:     string(job.Status),
		ErrorMsg:   job.ErrorMsg,
	}
}

func ParseStatusRequest(newStatus dto.JobStatusUpdateRequest) (*JobStatusUpdate, api_error.ApiErr) {
	jobStatusUpdate := JobStatusUpdate{}
	switch strings.ToLower(newStatus.Status) {
	case "created":
		jobStatusUpdate.newStatus = JobStatusCreated
	case "queued":
		jobStatusUpdate.newStatus = JobStatusQueued
	case "running":
		jobStatusUpdate.newStatus = JobStatusRunning
	case "paused":
		jobStatusUpdate.newStatus = JobStatusPaused
	case "finished":
		jobStatusUpdate.newStatus = JobStatusFinished
	case "failed":
		jobStatusUpdate.newStatus = JobStatusFailed
		jobStatusUpdate.errMsg = newStatus.ErrMsg
	default:
		return nil, api_error.NewBadRequestError(fmt.Sprintf("Could not parse status value %v", newStatus.Status))
	}
	return &jobStatusUpdate, nil
}
