package eurekago


type EurekaClientConfig struct {
	Username string
	Password string
	HeaderContentType string
	ServiceUrls []string   //url list of service
	RegisterWithEureka   bool   //register to eureka server
	AppName   string   //application name
	InstanceId    string  //instance id
	Port    int   //port
	HostName  string   //hostname
	RenewalIntervalInSecs  int  //lease-renewal-interval-in-seconds
	DurationInSecs   int  //lease-expiration-duration-in-seconds
}

//information of eureka instance
type InstanceInfo struct {
	Instance *WrapperInstanceInfo  `json:"instance"`
}

//information of eureka application
type ApplicationInfo struct {
	Application  *WrapperApplicationInfo  `json:"application"`
}

//information of eureka application list
type ApplicationList struct {
	Applications  *WrapperApplicationList  `json:"applications"`
}


type WrapperApplicationList struct {
	Version      string        `json:"versions__delta"`
	AppsHashcode string        `json:"versions__delta"`
	Applications []WrapperApplicationInfo `json:"application"`
}


type WrapperApplicationInfo struct {
	Name  string `json:"name"`
	Instance []WrapperInstanceInfo  `json:"instance"`
}

type WrapperInstanceInfo struct {
	InstanceId  string  `json:"instanceId,omitempty"`
	HostName string  `json:"hostName,omitempty"`
	App string  `json:"app,omitempty"`
	IpAddr string `json:"ipAddr,omitempty"`
	Status  string  `json:"status,omitempty"`
	OverriddenStatus  string  `json:"overriddenStatus,omitempty"`
	Port  *WrapperPort  `json:"port,omitempty"`
	SecurePort  *WrapperSecurePort  `json:"securePort,omitempty"`
	CountryId  int  `json:"countryId,omitempty"`
	DataCenterInfo  *WrapperDataCenterInfo   `json:"dataCenterInfo,omitempty"`
	LeaseInfo *WrapperLeaseInfo  `json:"leaseInfo,omitempty"`
	Metadata  *WrapperMetadata  `json:"metadata,omitempty"`
	HomePageUrl   string  `json:"homePageUrl,omitempty"`
	StatusPageUrl  string `json:"statusPageUrl,omitempty"`
	HealthCheckUrl  string `json:"healthCheckUrl,omitempty"`
	VipAddress  string  `json:"vipAddress,omitempty"`
	SecureVipAddress  string  `json:"secureVipAddress,omitempty"`
	IsCoordinatingDiscoveryServer  string  `json:"isCoordinatingDiscoveryServer,omitempty"`
	LastUpdatedTimestamp  string  `json:"lastUpdatedTimestamp,omitempty"`
	LastDirtyTimestamp  string  `json:"lastDirtyTimestamp,omitempty"`
	ActionType  string  `json:"actionType,omitempty"`
}


type WrapperPort struct {
	Port  int  `json:"$,omitempty"`
	Enabled interface{}  `json:"@enabled,omitempty"`
}


type WrapperSecurePort struct {
	Port  int  `json:"$,omitempty"`
	Enable  interface{}  `json:"@enabled"`
}

type WrapperDataCenterInfo struct {
	Class  string  `json:"@class,omitempty"`
	Name string  `json:"name,omitempty"`
}

type WrapperLeaseInfo struct {
	RenewalIntervalInSecs  int `json:"renewalIntervalInSecs,omitempty"`
	DurationInSecs  int  `json:"durationInSecs,omitempty"`
	RegistrationTimestamp  int64   `json:"registrationTimestamp,omitempty"`
	LastRenewalTimestamp  int64  `json:"lastRenewalTimestamp,omitempty"`
	EvictionTimestamp  int64  `json:"evictionTimestamp,omitempty"`
	ServiceUpTimestamp  int64   `json:"serviceUpTimestamp,omitempty"`
}

type WrapperMetadata struct {
	ManagementPort  string  `json:"management.port,omitempty"`
	JmxPort  string  `json:"jmx.port,omitempty"`
}