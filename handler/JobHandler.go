package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/johannes-kuhfuss/amssvc/dto"
	"github.com/johannes-kuhfuss/amssvc/service"
	"github.com/johannes-kuhfuss/services_utils/api_error"
	"github.com/johannes-kuhfuss/services_utils/logger"
	"github.com/microcosm-cc/bluemonday"
	"github.com/segmentio/ksuid"
)

type JobHandlers struct {
	Service service.JobService
}

var (
	policy *bluemonday.Policy
)

func init() {
	policy = bluemonday.UGCPolicy()
}

func getJobId(jobIdParam string) (string, api_error.ApiErr) {
	jobIdParam = policy.Sanitize(jobIdParam)
	jobId, err := ksuid.Parse(jobIdParam)
	if err != nil {
		logger.Error("User Id should be a ksuid", err)
		return "", api_error.NewBadRequestError("User id should be a ksuid")
	}
	return jobId.String(), nil
}

func (jh *JobHandlers) GetAllJobs(c *gin.Context) {
	status, _ := c.GetQuery("status")
	status = policy.Sanitize(status)
	jobs, err := jh.Service.GetAllJobs(status)
	if err != nil {
		logger.Error("Service error while getting all jobs", err)
		c.JSON(err.StatusCode(), err)
		return
	}
	c.JSON(http.StatusOK, jobs)
}

func (jh *JobHandlers) GetJobById(c *gin.Context) {
	jobId, err := getJobId(c.Param("job_id"))
	if err != nil {
		c.JSON(err.StatusCode(), err)
		return
	}
	job, err := jh.Service.GetJobById(jobId)
	if err != nil {
		logger.Error("Service error while getting job by id", err)
		c.JSON(err.StatusCode(), err)
		return
	}
	c.JSON(http.StatusOK, job)
}

func (jh *JobHandlers) CreateJob(c *gin.Context) {
	var newJobReq dto.NewJobRequest
	if err := c.ShouldBindJSON(&newJobReq); err != nil {
		logger.Error("invalid JSON body in create job request", err)
		apiErr := api_error.NewBadRequestError("invalid json body")
		c.JSON(apiErr.StatusCode(), apiErr)
		return
	}
	newJobReq.Name = policy.Sanitize(newJobReq.Name)
	newJobReq.SrcUrl = policy.Sanitize(newJobReq.SrcUrl)
	result, err := jh.Service.CreateJob(newJobReq)
	if err != nil {
		logger.Error("Service error while creating job", err)
		c.JSON(err.StatusCode(), err)
		return
	}
	c.JSON(http.StatusCreated, result)
}

func (jh JobHandlers) DeleteJobById(c *gin.Context) {
	jobId, err := getJobId(c.Param("job_id"))
	if err != nil {
		c.JSON(err.StatusCode(), err)
		return
	}
	err = jh.Service.DeleteJobById(jobId)
	if err != nil {
		logger.Error("Service error while deleting job by id", err)
		c.JSON(err.StatusCode(), err)
		return
	}
	c.JSON(http.StatusOK, nil)
}

func (jh JobHandlers) GetNextJob(c *gin.Context) {
	result, err := jh.Service.GetNextJob()
	if err != nil {
		logger.Error("Service error while getting next job", err)
		c.JSON(err.StatusCode(), err)
		return
	}
	c.JSON(http.StatusOK, result)
}
