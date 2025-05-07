package app

import "github.com/meteogo/weather-collector-service/internal/clients/open_meteo"

type Clients struct {
	openMeteoClient *open_meteo.Client
}

func InitClients() Clients {
	urlGenerator := open_meteo.NewURLGenerator()

	return Clients{
		openMeteoClient: open_meteo.NewOpenMeteoClient(urlGenerator),
	}
}
