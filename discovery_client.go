package eurekago

import (
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

var (
	paramError = errors.New("parameter error")
)

func NewDiscoveryClient(conf *EurekaClientConfig) (registered bool, discovery DiscoveryClient, err error) {
	if conf.AppName == "" || conf.InstanceId == "" || len(conf.ServiceUrls) == 0 {
		err = paramError
		return
	}

	isJson := getIsJsonContentType(conf)

	wrapperInstance := getWrapperInstanceInfo(conf)
	eurekaClient := newEurekaHttpClient(conf.ServiceUrls, conf.Username, conf.Password, isJson)

	disClient := &discoveryClient{
		eurekaClient:       eurekaClient,
		isJsonContentType:  isJson,
		serviceUrls:        conf.ServiceUrls,
		instanceConf:       wrapperInstance,
		lastDirtyTimestamp: wrapperInstance.LastDirtyTimestamp,
		closeChan:          make(chan struct{}),
	}
	discovery = disClient

	//register to eureka service
	if conf.RegisterWithEureka {
		registered, err = eurekaClient.Register(&InstanceInfo{wrapperInstance})
		go disClient.heartbeatTask()
	}

	return
}


func getIsJsonContentType(conf *EurekaClientConfig) bool {
	return  conf.HeaderContentType == ""  || strings.Contains(strings.ToUpper(conf.HeaderContentType),"JSON")
}

type discoveryClient struct {
	eurekaClient       EurekaHttpClient
	isJsonContentType  bool
	urlIndex           int
	serviceUrls        []string
	instanceConf       *WrapperInstanceInfo
	isDirty            bool
	lastDirtyTimestamp string
	wg                 sync.WaitGroup
	closeChan          chan struct{}
}

type DiscoveryClient interface {
	DiscoveryStatusUpdate(status  string) (bool, error)
	GetApplications(regions ...string) (*ApplicationList, error)
	GetApplication(appName string) (*ApplicationInfo, error)
	GetInstance(appName, instanceId string) (*InstanceInfo, error)
	GetInstanceById(instanceId string) (*InstanceInfo, error)
	Shutdown()
}

//only update myself
func (d *discoveryClient)DiscoveryStatusUpdate(status  string) (bool, error) {
	return d.eurekaClient.StatusUpdate(d.instanceConf.App,d.instanceConf.InstanceId,status,d.lastDirtyTimestamp)
}

//if regions is nil, get all application's information
//if regions is not nil,get application's information by regions
func (d *discoveryClient)GetApplications(regions ...string) (*ApplicationList, error)  {
	return d.eurekaClient.GetApplications(regions...)
}

//get application's information by appName
func (d *discoveryClient)GetApplication(appName string) (*ApplicationInfo, error)  {
	return d.eurekaClient.GetApplication(appName)
}

//get instance's information by appName and instanceId
func (d *discoveryClient)GetInstance(appName, instanceId string) (*InstanceInfo, error) {
	return d.eurekaClient.GetInstance(appName,instanceId)
}

//get instance's information by instanceId
func (d *discoveryClient)GetInstanceById(instanceId string) (*InstanceInfo, error)  {
	return d.eurekaClient.GetInstanceById(instanceId)
}


func (d *discoveryClient) Shutdown()  {
	close(d.closeChan)
	_, _ = d.eurekaClient.Deregister(d.instanceConf.App,d.instanceConf.InstanceId)
	d.wg.Wait()
}

func (d *discoveryClient) heartbeatTask() {
	var (
		interval = d.instanceConf.LeaseInfo.RenewalIntervalInSecs
		ticker = time.NewTicker(time.Second* time.Duration(interval))
		err error
	)

	d.wg.Add(1)

	for {
		select {
		case <- ticker.C:
			_,err = d.sendHeartbeat()
			if err != nil {
				log.Println("send heartbeat error, ",err.Error())
			}

		case <- d.closeChan:
			goto _exit
		}
	}

	_exit:
		log.Println("exit heartbeat task")
		d.wg.Done()
}


//send heartbeat to eureka service
func (d *discoveryClient) sendHeartbeat() (success bool, err error){
	var statusCode int
	statusCode,err = d.eurekaClient.SendHeartBeat(d.instanceConf.App,d.instanceConf.InstanceId,statusUp,d.lastDirtyTimestamp,"")
	if err != nil {
		if len(d.serviceUrls) > 1 {
			d.urlIndex = (d.urlIndex + 1) % len(d.serviceUrls)
		}
		return
	}

	switch statusCode {
	case http.StatusNoContent,http.StatusOK:
		success = true
	case http.StatusNotFound:
		//instance not found,then register
		d.setIsDirty()
		success,_ = d.eurekaClient.Register(&InstanceInfo{d.instanceConf})
		if success {
			d.unsetIsDirty()
		}
	}

	return
}

func (d *discoveryClient) setIsDirty()  {
	d.isDirty = true
	d.lastDirtyTimestamp = fmt.Sprintf("%d",time.Now().UnixNano()/1000000)
}


func (d *discoveryClient) unsetIsDirty() {
	d.isDirty = false
}


func getWrapperInstanceInfo(conf *EurekaClientConfig) *WrapperInstanceInfo {
	renewalIntervalInSecs := defaultLeaseRenewalInterval
	durationInSecs := defaultLeaseDuration
	if conf.RenewalIntervalInSecs > 0 {
		renewalIntervalInSecs = conf.RenewalIntervalInSecs
	}
	if conf.DurationInSecs > 0 {
		durationInSecs = conf.DurationInSecs
	}

	hostName := defaultHostName
	if conf.HostName != "" {
		hostName = conf.HostName
	}

	currentTimeStr := fmt.Sprintf("%d",time.Now().UnixNano()/1000000)

	return &WrapperInstanceInfo{
		InstanceId : conf.InstanceId,
		HostName  : hostName,
		App : conf.AppName,
		IpAddr : getLocalIP(),
		Status: statusUp,
		Port : &WrapperPort{
			Port   : conf.Port,
			Enabled: true,
		},
		DataCenterInfo: &WrapperDataCenterInfo{
			Class  : defaultDataCenterInfoClass,
			Name  : defaultDataCenterInfoName,
		},
		LeaseInfo : &WrapperLeaseInfo{
			RenewalIntervalInSecs  : renewalIntervalInSecs,
			DurationInSecs  : durationInSecs,
		},
		Metadata : &WrapperMetadata{
			ManagementPort : fmt.Sprintf("%d",conf.Port),
		},
		VipAddress: conf.AppName,
		SecureVipAddress: conf.AppName,
		LastUpdatedTimestamp: currentTimeStr,
		LastDirtyTimestamp: currentTimeStr,
	}
}

//get local ip address
func getLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}

	for i := range addrs {
		// address type  not a loopback address
		if ipNet, ok := addrs[i].(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
			if ipNet.IP.To4() != nil {
				return ipNet.IP.String()
			}
		}
	}

	return ""
}