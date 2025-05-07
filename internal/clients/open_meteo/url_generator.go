package open_meteo

import (
	"fmt"
	"sort"
	"strings"

	"github.com/meteogo/weather-collector-service/internal/pkg/enums"
	"github.com/meteogo/weather-collector-service/internal/services/weather_service"
)

type urlGeneratorImpl struct {
}

func NewURLGenerator() *urlGeneratorImpl {
	return &urlGeneratorImpl{}
}

func (g *urlGeneratorImpl) GenerateURL(coordinates weather_service.Coordinates, params weather_service.MonitoringParamsMap) string {
	var (
		baseURL       = "https://api.open-meteo.com/v1/forecast"
		latParam      = fmt.Sprintf("latitude=%.2f", coordinates.Lat)
		longParam     = fmt.Sprintf("longitude=%.2f", coordinates.Long)
		currentParams = make([]string, 0, len(params))
	)

	keys := make([]enums.MonitoringParam, 0, len(params))
	for key := range params {
		keys = append(keys, key)
	}

	sort.Slice(keys, func(i, j int) bool {
		return string(keys[i]) < string(keys[j])
	})

	for _, key := range keys {
		currentParams = append(currentParams, params[key])
	}

	currentParamStr := fmt.Sprintf("current=%s", strings.Join(currentParams, ","))
	queryParams := []string{latParam, longParam, currentParamStr}
	return fmt.Sprintf("%s?%s", baseURL, strings.Join(queryParams, "&"))
}
