package enums

type MonitoringParam string

const (
	MonitoringParamTemperature      = MonitoringParam("temperature")
	MonitoringParamRelativeHumidity = MonitoringParam("relativeHumidity")
	MonitoringParamWindSpeed        = MonitoringParam("windSpeed")
	MonitoringParamWeatherCode      = MonitoringParam("weatherCode")
	MonitoringParamCloudCover       = MonitoringParam("cloudCover")
	MonitoringParamPrecipitation    = MonitoringParam("precipitation")
	MonitoringParamVisibility       = MonitoringParam("visibility")
)
