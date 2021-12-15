package app

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/johannes-kuhfuss/amssvc/config"
	"github.com/johannes-kuhfuss/amssvc/domain"
	"github.com/johannes-kuhfuss/amssvc/handler"
	"github.com/johannes-kuhfuss/amssvc/service"
	"github.com/johannes-kuhfuss/services_utils/logger"
)

var (
	router     *gin.Engine
	jobHandler handler.JobHandlers
	jobService service.JobService
)

func initRouter() {
	gin.SetMode(config.GinMode)
	gin.DefaultWriter = logger.GetLogger()
	router = gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
}

func wireApp() {
	customerRepo := domain.NewJobRepositoryMem()
	jobService = service.NewJobService(customerRepo)
	jobHandler = handler.JobHandlers{Service: jobService}
}

func startRouter() {
	listenAddr := fmt.Sprintf("%s:%s", config.ServerAddr, config.ServerPort)
	logger.Info(fmt.Sprintf("Listening on %v", listenAddr))
	if err := router.Run(listenAddr); err != nil {
		logger.Error("Error while starting router", err)
		panic(err)
	}
}

func StartApp() {
	logger.Info("Starting application")
	err := config.InitConfig(config.EnvFile)
	if err != nil {
		panic(err)
	}

	initRouter()
	wireApp()
	mapUrls()
	startProcessing()
	startRouter()
	logger.Info("Application ended")
}

func startProcessing() {
}
