package eurekago

//default config
const (
	defaultLeaseRenewalInterval = 30
	defaultLeaseDuration  = 90
	defaultServiceUrl  = "http://127.0.0.1:8761/eureka/"
	defaultRegisterWithEureka = true
	defaultHostName = "localhost"
	defaultDataCenterInfoClass = "com.netflix.appinfo.InstanceInfo$DefaultDataCenterInfo"
	defaultDataCenterInfoName = "MyOwn"
)

//content-type
const (
	applicationJsonValue = "application/json"
	applicationJsonUtf8Value = "application/json;charset=UTF-8"
	applicationXmlValue = "application/xml"
)

//eureka status
const (
	statusUp = "UP"  // Ready to receive traffic
	statusDown = "DOWN" // Do not send traffic- healthcheck callback failed
	statusStarting = "STARTING"  //Just about starting- initializations to be done - do not
	statusOutOfService = "OUT_OF_SERVICE"  //Intentionally shutdown for traffic
	statusUnknown = "UNKNOWN"
)