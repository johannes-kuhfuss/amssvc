package domain

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/johannes-kuhfuss/services_utils/api_error"
	"github.com/johannes-kuhfuss/services_utils/date"
)

type JobRepositoryMem struct {
	jobList map[string]Job
	mu      *sync.Mutex
}

func NewJobRepositoryMem() JobRepositoryMem {
	jList := make(map[string]Job)
	m := sync.Mutex{}
	return JobRepositoryMem{jList, &m}
}

func (jrm JobRepositoryMem) FindAll(status string) (*[]Job, api_error.ApiErr) {
	jrm.mu.Lock()
	defer jrm.mu.Unlock()
	if len(jrm.jobList) == 0 {
		return nil, api_error.NewNotFoundError("no jobs in joblist")
	}
	if strings.TrimSpace(status) == "" {
		return convertMapToSlice(jrm.jobList), nil
	} else {
		return filterByStatus(jrm.jobList, status)
	}
}

func convertMapToSlice(jList map[string]Job) *[]Job {
	slice := make([]Job, 0)
	for _, job := range jList {
		slice = append(slice, job)
	}
	return &slice
}

func filterByStatus(jList map[string]Job, status string) (*[]Job, api_error.ApiErr) {
	filteredByStatusList := make([]Job, 0)
	for _, curJob := range jList {
		if curJob.Status == JobStatus(status) {
			filteredByStatusList = append(filteredByStatusList, curJob)
		}
	}
	if len(filteredByStatusList) == 0 {
		return nil, api_error.NewNotFoundError(fmt.Sprintf("no jobs with status %v in joblist", status))
	} else {
		return &filteredByStatusList, nil
	}
}

func (csm JobRepositoryMem) FindById(id string) (*Job, api_error.ApiErr) {
	csm.mu.Lock()
	defer csm.mu.Unlock()
	if len(csm.jobList) == 0 {
		return nil, api_error.NewNotFoundError("no jobs in joblist")
	}
	return filterById(csm.jobList, id)
}

func filterById(jList map[string]Job, id string) (*Job, api_error.ApiErr) {
	for _, curJob := range jList {
		if curJob.Id.String() == id {
			return &curJob, nil
		}
	}
	return nil, api_error.NewNotFoundError(fmt.Sprintf("no job with id %v in joblist", id))
}

func (csm JobRepositoryMem) Save(job Job) api_error.ApiErr {
	csm.mu.Lock()
	defer csm.mu.Unlock()
	job.ModifiedAt = date.GetNowUtc()
	csm.jobList[job.Id.String()] = job
	return nil
}

func (csm JobRepositoryMem) DeleteById(id string) api_error.ApiErr {
	csm.mu.Lock()
	defer csm.mu.Unlock()
	if len(csm.jobList) == 0 {
		return api_error.NewNotFoundError("no jobs in joblist")
	}
	_, err := filterById(csm.jobList, id)
	if err != nil {
		return err
	}
	delete(csm.jobList, id)
	return nil
}

func (csm JobRepositoryMem) GetNext() (*Job, api_error.ApiErr) {
	var nextJobId string = ""
	var nextJobDate time.Time = date.GetNowUtc().Add(1 * time.Second)

	csm.mu.Lock()
	defer csm.mu.Unlock()

	if len(csm.jobList) == 0 {
		err := api_error.NewNotFoundError("no jobs in joblist")
		return nil, err
	}
	for _, job := range csm.jobList {
		if job.Status == JobStatusCreated {
			if job.CreatedAt.Before(nextJobDate) {
				nextJobDate = job.CreatedAt
				nextJobId = job.Id.String()
			}
		}
	}
	if nextJobId == "" {
		err := api_error.NewNotFoundError("no jobs with status created in joblist")
		return nil, err
	}
	job, _ := filterById(csm.jobList, nextJobId)
	return job, nil
}

func (csm JobRepositoryMem) SetStatus(id string, newStatus JobStatusUpdate) api_error.ApiErr {
	job, err := csm.FindById(id)
	if err != nil {
		return err
	}
	job.Status = newStatus.newStatus
	job.ErrorMsg = newStatus.errMsg
	csm.Save(*job)
	return nil
}
