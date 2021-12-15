package config

import (
	"os"

	"github.com/gin-gonic/gin"
	"github.com/johannes-kuhfuss/services_utils/logger"
	"github.com/joho/godotenv"
)

const (
	EnvFile = ".env"
)

var (
	GinMode       string
	ServerAddr    string
	ServerPort    string
	Shutdown      bool = false
	NoJobWaitTime int  = 10
)

func InitConfig(file string) error {
	logger.Info("Initalizing configuration")
	loadConfig(file)
	configGin()
	configServer()
	logger.Info("Done initalizing configuration")
	return nil
}

func loadConfig(file string) error {
	err := godotenv.Load(file)
	if err != nil {
		logger.Error("Could not open env file", err)
		return err
	}
	return nil
}

func configGin() {
	ginMode, ok := os.LookupEnv("GIN_MODE")
	if !ok || (ginMode != gin.ReleaseMode && ginMode != gin.DebugMode && ginMode != gin.TestMode) {
		GinMode = "release"
	} else {
		GinMode = ginMode
	}
}

func configServer() {
	var ok bool
	ServerAddr, ok = os.LookupEnv("SERVER_ADDR")
	if !ok {
		ServerAddr = ""
	}
	ServerPort, ok = os.LookupEnv("SERVER_PORT")
	if !ok {
		ServerPort = "8080"
	}
}
