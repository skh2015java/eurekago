package eurekago

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"net/http"
	"strings"
)

var (
	uriApps           = "/apps/"
	headerAccept      = "Accept"
	headerContentType = "Content-Type"
)

type EurekaHttpClient interface {
	Register(info *InstanceInfo) (bool, error)
	Deregister(appName, instanceId string) (bool, error)
	SendHeartBeat(appName, instanceId, status, lastDirtyTimestamp, overriddenStatus string) (statusCode int, err error)
	StatusUpdate(appName, instanceId, status, lastDirtyTimestamp string) (bool, error)
	GetApplications(regions ...string) (*ApplicationList, error)
	GetApplication(appName string) (*ApplicationInfo, error)
	GetInstance(appName, instanceId string) (*InstanceInfo, error)
	GetInstanceById(instanceId string) (*InstanceInfo, error)
}

type httpClient struct {
	isJson             bool
	serviceUrl         string
	urlIndex           int
	serviceUrlList   []string
	username, password string
	headers            map[string]string
}

func newEurekaHttpClient(serviceUrlList []string, username, password string, contentIsJson bool) EurekaHttpClient {
	headerMap := make(map[string]string)
	if contentIsJson {
		headerMap[headerAccept] = applicationJsonValue
		headerMap[headerContentType] = applicationJsonValue
	} else {
		headerMap[headerAccept] = applicationXmlValue
		headerMap[headerContentType] = applicationXmlValue
	}

	for i := range serviceUrlList {
		serviceUrlList[i] = strings.TrimRight(serviceUrlList[i],"/")
	}

	return &httpClient{
		isJson:     contentIsJson,
		serviceUrl: serviceUrlList[0],
		serviceUrlList: serviceUrlList,
		username:   username,
		password:   password,
		headers:    headerMap,
	}
}

//register to eureka server
func (hc *httpClient) Register(info *InstanceInfo) (success bool, err error) {
	var body []byte
	body, err = json.Marshal(info)
	if err != nil {
		return
	}
	param := &requestParam{
		URL:      hc.serviceUrl + uriApps + info.Instance.App,
		Method:   http.MethodPost,
		Headers:  hc.headers,
		Body:     string(body),
		Username: hc.username,
		Password: hc.password,
	}
	var statusCode int
	_, statusCode, err = handleHttpRequest(param)
	if err != nil {
		hc.handleError(err)
	}
	if statusCode == http.StatusOK || statusCode == http.StatusNoContent {
		success = true
	}

	return
}

//deregister from eureka server
func (hc *httpClient) Deregister(appName, instanceId string) (bool, error) {
	param := &requestParam{
		URL:      hc.serviceUrl + uriApps + appName + "/" + instanceId,
		Method:   http.MethodDelete,
		Headers:  hc.headers,
		Username: hc.username,
		Password: hc.password,
	}
	_, statusCode, err := handleHttpRequest(param)
	if err != nil {
		hc.handleError(err)
	}
	return statusCode == http.StatusOK, err
}

//send heartbeat to eureka
func (hc *httpClient) SendHeartBeat(appName, instanceId, status, lastDirtyTimestamp, overriddenStatus string) (statusCode int, err error) {
	url := hc.serviceUrl + uriApps + appName + "/" + instanceId + "?status=" + status + "&lastDirtyTimestamp=" + lastDirtyTimestamp
	if overriddenStatus != "" {
		url += "&overriddenstatus=" + overriddenStatus
	}
	param := &requestParam{
		URL:      url,
		Method:   http.MethodPut,
		Headers:  hc.headers,
		Username: hc.username,
		Password: hc.password,
	}
	_, statusCode, err = handleHttpRequest(param)
	if err != nil {
		hc.handleError(err)
	}
	return
}

//update eureka client status
func (hc *httpClient) StatusUpdate(appName, instanceId, status, lastDirtyTimestamp string) (bool, error) {
	param := &requestParam{
		URL:      hc.serviceUrl + uriApps + appName + "/" + instanceId + "/status?value=" + status + "&lastDirtyTimestamp=" + lastDirtyTimestamp,
		Method:   http.MethodPut,
		Headers:  hc.headers,
		Username: hc.username,
		Password: hc.password,
	}

	_, statusCode, err := handleHttpRequest(param)
	if err != nil {
		hc.handleError(err)
	}

	return statusCode == http.StatusOK, err
}

//if regions is nil, get all application's information
//if regions is not nil,get application's information by regions
func (hc *httpClient) GetApplications(regions ...string) (*ApplicationList, error) {
	url := hc.serviceUrl + uriApps
	if len(regions) > 0 {
		url += "?regions=" + strings.Join(regions, ",")
	}

	param := &requestParam{
		URL:      url,
		Method:   http.MethodGet,
		Headers:  hc.headers,
		Username: hc.username,
		Password: hc.password,
	}

	var result ApplicationList
	err := hc.requestAndFormatResult(param, &result)

	return &result, err
}

//get application's information by appName
func (hc *httpClient) GetApplication(appName string) (*ApplicationInfo, error) {
	param := &requestParam{
		URL:      hc.serviceUrl + uriApps + appName,
		Method:   http.MethodGet,
		Headers:  hc.headers,
		Username: hc.username,
		Password: hc.password,
	}

	var result ApplicationInfo
	err := hc.requestAndFormatResult(param, &result)

	return &result, err
}

//get instance's information by appName and instanceId
func (hc *httpClient) GetInstance(appName, instanceId string) (*InstanceInfo, error) {
	param := &requestParam{
		URL:      hc.serviceUrl + uriApps + appName + "/" + instanceId,
		Method:   http.MethodGet,
		Headers:  hc.headers,
		Username: hc.username,
		Password: hc.password,
	}

	var result InstanceInfo
	err := hc.requestAndFormatResult(param, &result)

	return &result, err
}

//get instance's information by instanceId
func (hc *httpClient) GetInstanceById(instanceId string) (*InstanceInfo, error) {
	param := &requestParam{
		URL:      hc.serviceUrl + "/instances/" + instanceId,
		Method:   http.MethodGet,
		Headers:  hc.headers,
		Username: hc.username,
		Password: hc.password,
	}

	var result InstanceInfo
	err := hc.requestAndFormatResult(param, &result)

	return &result, err
}

func (hc *httpClient) requestAndFormatResult(param *requestParam, result interface{}) (err error) {
	var respBody []byte
	respBody, _, err = handleHttpRequest(param)
	if err != nil {
		hc.handleError(err)
		return
	}
	if len(respBody) == 0 {
		err = errors.New("data empty")
		return
	}
	if hc.isJson {
		err = json.Unmarshal(respBody, result)
	} else {
		err = xml.Unmarshal(respBody, result)
	}

	return
}

func (hc *httpClient) handleError(err error) {
	if len(hc.serviceUrlList) > 1 {
		hc.urlIndex = (hc.urlIndex + 1) % len(hc.serviceUrlList)
		hc.serviceUrl = hc.serviceUrlList[hc.urlIndex]
	}
}