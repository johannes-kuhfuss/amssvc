package app

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/johannes-kuhfuss/amssvc/config"
	"github.com/johannes-kuhfuss/amssvc/domain"
	"github.com/johannes-kuhfuss/amssvc/dto"
	"github.com/johannes-kuhfuss/amssvc/handler"
	"github.com/johannes-kuhfuss/amssvc/service"
	"github.com/johannes-kuhfuss/services_utils/api_error"
	"github.com/johannes-kuhfuss/services_utils/logger"
)

var (
	router     *gin.Engine
	jobHandler handler.JobHandlers
	jobService service.JobService
	authToken  string
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
	token, err := getAzureAuthToken()
	if err != nil {
		panic(err)
	}
	authToken = token
	go processItems()
}

func processItems() {
	transformName := "jkuencodeproxy"
	apiVersion := "2020-05-01"
	for !config.Shutdown {
		job, err := jobService.GetNextJob()
		if err != nil {
			logger.Debug(err.Message())
			time.Sleep(time.Second * time.Duration(config.NoJobWaitTime))
		} else {
			postUrl := fmt.Sprintf("%v/subscriptions/%v/resourceGroups/%v/providers/Microsoft.Media/mediaServices/%v/transforms/%v/jobs/%v?api-version=%v",
				config.ArmEndpoint, config.SubscriptionId, config.ResourceGroup, config.AccountName, transformName, job.Id, apiVersion)
			_ = postUrl
			url, _ := url.Parse(job.SrcUrl)
			pathName := strings.TrimLeft(filepath.Dir(url.Path), string(os.PathSeparator))
			baseUrl := url.Scheme + "://" + url.Host + "/" + pathName + "/"
			fileName := filepath.Base(job.SrcUrl)
			files := make([]string, 1)
			files = append(files, fileName)

			outputs := make([]dto.Outputs, 1)
			output := dto.Outputs{
				OdataType: "#Microsoft.Media.JobOutputAsset",
				AssetName: fmt.Sprintf("%v_proxy", fileName),
			}
			outputs = append(outputs, output)
			amsJobReq := dto.AmsJobReq{
				Properties: dto.Properties{
					Input: dto.Input{
						OdataType: "#Microsoft.Media.JobInputHttp",
						BaseURI:   baseUrl,
						Files:     files,
					},
					Outputs:  outputs,
					Priority: "Normal",
				},
			}
			fmt.Printf("Request %v\n", amsJobReq)
			time.Sleep(time.Second * time.Duration(config.NoJobWaitTime))
		}
	}
}

func getAzureAuthToken() (string, api_error.ApiErr) {
	postUrl := fmt.Sprintf("https://login.microsoftonline.com/%v/oauth2/token", config.AadTenantDomain)
	postData := url.Values{
		"grant_type":    {"client_credentials"},
		"client_id":     {config.AadClientId},
		"client_secret": {config.AadSecret},
		"resource":      {config.ArmAadAudience},
	}
	resp, err := http.PostForm(postUrl, postData)
	if err != nil {
		logger.Error("Error while getting auth token.", err)
		return "", api_error.NewInternalServerError("Error while getting auth token.", err)
	}
	var result = make(map[string]string)
	json.NewDecoder(resp.Body).Decode(&result)
	if result["access_token"] == "" {
		logger.Error("Error while getting auth token.", nil)
		return "", api_error.NewInternalServerError("Error while getting auth token.", nil)
	}
	return result["access_token"], nil
}
