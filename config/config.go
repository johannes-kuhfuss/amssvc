package config

import (
	"errors"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/johannes-kuhfuss/services_utils/logger"
	"github.com/joho/godotenv"
)

const (
	EnvFile = ".env"
)

var (
	GinMode         string
	ServerAddr      string
	ServerPort      string
	Shutdown        bool = false
	NoJobWaitTime   int  = 10
	AadClientId     string
	AadSecret       string
	AadTenantDomain string
	AadTenantId     string
	AccountName     string
	ResourceGroup   string
	SubscriptionId  string
	ArmAadAudience  string
	ArmEndpoint     string
)

func InitConfig(file string) error {
	logger.Info("Initalizing configuration")
	loadConfig(file)
	configGin()
	err := configMediaServices()
	if err != nil {
		return err
	}
	configServer()
	logger.Info("Done initalizing configuration")
	return nil
}

func configMediaServices() error {
	var ok bool
	AadClientId, ok = os.LookupEnv("AADCLIENTID")
	if !ok || AadClientId == "" {
		return errors.New("no AadClientId set")
	}
	AadSecret, ok = os.LookupEnv("AADSECRET")
	if !ok || AadSecret == "" {
		return errors.New("no AadSecret set")
	}
	AadTenantDomain, ok = os.LookupEnv("AADTENANTDOMAIN")
	if !ok || AadTenantDomain == "" {
		return errors.New("no AadTenantDomain set")
	}
	AadTenantId, ok = os.LookupEnv("AADTENANTID")
	if !ok || AadTenantId == "" {
		return errors.New("no AadTenantId set")
	}
	AccountName, ok = os.LookupEnv("ACCOUNTNAME")
	if !ok || AccountName == "" {
		return errors.New("no AccountName set")
	}
	ResourceGroup, ok = os.LookupEnv("RESOURCEGROUP")
	if !ok || ResourceGroup == "" {
		return errors.New("no ResourceGroup set")
	}
	SubscriptionId, ok = os.LookupEnv("SUBSCRIPTIONID")
	if !ok || SubscriptionId == "" {
		return errors.New("no SubscriptionId set")
	}
	ArmAadAudience, ok = os.LookupEnv("ARMAADAUDIENCE")
	if !ok || ArmAadAudience == "" {
		return errors.New("no ArmAadAudience set")
	}
	ArmEndpoint, ok = os.LookupEnv("ARMENDPOINT")
	if !ok || ArmEndpoint == "" {
		return errors.New("no ArmEndpoint set")
	}
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
