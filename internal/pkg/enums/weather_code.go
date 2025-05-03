package enums

type WeatherCode int32

const (
	ClearSky               WeatherCode = 0
	MainlyClear            WeatherCode = 1
	PartlyCloudy           WeatherCode = 2
	Overcast               WeatherCode = 3
	Fog                    WeatherCode = 45
	DepositingRimeFog      WeatherCode = 48
	DrizzleLight           WeatherCode = 51
	DrizzleModerate        WeatherCode = 53
	DrizzleDense           WeatherCode = 55
	FreezingDrizzleLight   WeatherCode = 56
	FreezingDrizzleDense   WeatherCode = 57
	RainSlight             WeatherCode = 61
	RainModerate           WeatherCode = 63
	RainHeavy              WeatherCode = 65
	FreezingRainLight      WeatherCode = 66
	FreezingRainHeavy      WeatherCode = 67
	SnowFallSlight         WeatherCode = 71
	SnowFallModerate       WeatherCode = 73
	SnowFallHeavy          WeatherCode = 75
	SnowGrains             WeatherCode = 77
	RainShowersSlight      WeatherCode = 80
	RainShowersModerate    WeatherCode = 81
	RainShowersViolent     WeatherCode = 82
	SnowShowersSlight      WeatherCode = 85
	SnowShowersHeavy       WeatherCode = 86
	ThunderstormSlight     WeatherCode = 95
	ThunderstormHailSlight WeatherCode = 96
	ThunderstormHailHeavy  WeatherCode = 99
)
