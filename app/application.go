package app

import (
	"bytes"
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
	router        *gin.Engine
	jobHandler    handler.JobHandlers
	jobService    service.JobService
	authToken     string
	transformName string = "jkuencodeproxy"
	apiVersion    string = "2020-05-01"
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
	go renewAzureAuthToken()
	go processItems()
}

func renewAzureAuthToken() {
	for !config.Shutdown {
		token, err := getAzureAuthToken()
		if err != nil {
			logger.Error("", err)
			time.Sleep(time.Second * 30)
		} else {
			authToken = token
			logger.Info("set new auth token")
			time.Sleep(time.Second * 3500)
		}
	}
}

func processItems() {
	for !config.Shutdown {
		job, err := jobService.GetNextJob()
		if err != nil {
			logger.Debug(err.Message())
			time.Sleep(time.Second * time.Duration(config.NoJobWaitTime))
		} else {
			createDestAsset(job)
			createProxy(job)
			newStatus := dto.JobStatusUpdateRequest{
				Status: "finished",
				ErrMsg: "",
			}
			jobService.SetStatus(job.Id, newStatus)
		}
	}
}

func createDestAsset(job *dto.JobResponse) {
	bearer := "Bearer " + authToken
	putUrl := createAssetRequestUrl(job.SrcUrl)
	assetReq := dto.AssetReq{}
	reqJson, _ := json.Marshal(assetReq)
	client := &http.Client{}
	req, err := http.NewRequest(http.MethodPut, putUrl, bytes.NewBuffer(reqJson))
	if err != nil {
		logger.Error("error creating asset", err)
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Add("Authorization", bearer)
	_, err = client.Do(req)
	if err != nil {
		logger.Error("error creating asset", err)
	}
	logger.Info(fmt.Sprintf("Created new asset for %v", job.SrcUrl))
}

func createAssetRequestUrl(srcUrl string) string {
	fileName := filepath.Base(srcUrl)
	fileNameNoExt := strings.TrimSuffix(fileName, filepath.Ext(fileName))
	assetName := fileNameNoExt + "_proxy"
	return fmt.Sprintf("%v/subscriptions/%v/resourceGroups/%v/providers/Microsoft.Media/mediaServices/%v/assets/%v?api-version=%v",
		config.ArmEndpoint, config.SubscriptionId, config.ResourceGroup, config.AccountName, assetName, apiVersion)
}

func createProxy(job *dto.JobResponse) {
	bearer := "Bearer " + authToken
	putUrl := createJobRequestUrl(job.Id)
	amsJobReq := createAmsJobRequest(job.SrcUrl)
	reqJson, _ := json.Marshal(amsJobReq)
	client := &http.Client{}
	req, err := http.NewRequest(http.MethodPut, putUrl, bytes.NewBuffer(reqJson))
	if err != nil {
		logger.Error("error creating proxy", err)
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Add("Authorization", bearer)
	_, err = client.Do(req)
	if err != nil {
		logger.Error("error creating proxy", err)
	}
	logger.Info(fmt.Sprintf("Created new proxy for %v", job.SrcUrl))
}

func createAmsJobRequest(srcUrl string) dto.AmsJobReq {
	var files []string
	var outputs []dto.Outputs
	url, _ := url.Parse(srcUrl)
	pathName := strings.TrimLeft(filepath.Dir(url.Path), string(os.PathSeparator))
	baseUrl := url.Scheme + "://" + url.Host + "/" + pathName + "/" + config.SasToken
	fileName := filepath.Base(srcUrl)
	fileNameNoExt := strings.TrimSuffix(fileName, filepath.Ext(fileName))
	files = append(files, fileName)

	output := dto.Outputs{
		OdataType: "#Microsoft.Media.JobOutputAsset",
		AssetName: fmt.Sprintf("%v_proxy", fileNameNoExt),
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
	return amsJobReq
}

func createJobRequestUrl(jobId string) string {
	return fmt.Sprintf("%v/subscriptions/%v/resourceGroups/%v/providers/Microsoft.Media/mediaServices/%v/transforms/%v/jobs/%v?api-version=%v",
		config.ArmEndpoint, config.SubscriptionId, config.ResourceGroup, config.AccountName, transformName, jobId, apiVersion)
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
