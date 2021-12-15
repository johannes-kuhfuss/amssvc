package app

func mapUrls() {
	router.GET("/jobs", jobHandler.GetAllJobs)
	router.GET("jobs/:job_id", jobHandler.GetJobById)
	router.POST("/jobs", jobHandler.CreateJob)
	router.DELETE("jobs/:job_id", jobHandler.DeleteJobById)
	router.GET("/jobs/next", jobHandler.GetNextJob)
}
