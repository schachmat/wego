package backends

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/schachmat/wego/iface"
)

type forecastConfig struct {
	apiKey   string
	language string
	debug    bool
}

type forecastDataPoint struct {
	Time                *float64 `json:"time"`
	Summary             string   `json:"summary"`
	Icon                string   `json:"icon"`
	SunriseTime         *float32 `json:"sunriseTime"`
	SunsetTime          *float32 `json:"sunsetTime"`
	PrecipIntensity     *float32 `json:"precipIntensity"`
	PrecipProbability   *float32 `json:"precipProbability"`
	Temperature         *float32 `json:"temperature"`
	TemperatureMin      *float32 `json:"temperatureMin"`
	TemperatureMax      *float32 `json:"temperatureMax"`
	ApparentTemperature *float32 `json:"apparentTemperature"`
	WindSpeed           *float32 `json:"windSpeed"`
	WindBearing         *float32 `json:"windBearing"`
	Visibility          *float32 `json:"visibility"`
}

type forecastDataBlock struct {
	Summary string              `json:"summary"`
	Icon    string              `json:"icon"`
	Data    []forecastDataPoint `json:"data"`
}

type forecastResponse struct {
	Latitude  *float32          `json:"latitude"`
	Longitude *float32          `json:"longitude"`
	Timezone  string            `json:"timezone"`
	Offset    *float32          `json:"offset"`
	Currently forecastDataPoint `json:"currently"`
	Hourly    forecastDataBlock `json:"hourly"`
	Code      *int              `json:"code"`
}

const (
	// see https://developer.forecast.io/docs/v2
	// see also https://github.com/mlbright/forecast
	//https://api.forecast.io/forecast/APIKEY/LATITUDE,LONGITUDE
	forecastWuri = "https://api.forecast.io/forecast/%s/%s?units=ca&lang=%s&exclude=minutely,daily,alerts,flags&extend=hourly"
)

func forecastParseDaily(db forecastDataBlock, numdays int) []iface.Day {
	var forecast []iface.Day
	var day *iface.Day

	//TODO: fill current day forecast with time machine call

	for _, dp := range db.Data {
		slot, err := forecastParseCond(dp)
		if err != nil {
			log.Println("Error parsing hourly weather condition:", err)
			continue
		}

		if day != nil && day.Date.Day() != slot.Time.Day() {
			if len(forecast) >= numdays-1 {
				break
			}
			forecast = append(forecast, *day)
			day = nil
		}
		if day == nil {
			day = new(iface.Day)
			day.Date = slot.Time
			//TODO: min-,max-temperature, astronomy
		}

		day.Slots = append(day.Slots, slot)
	}
	return append(forecast, *day)
}

func forecastParseCond(dp forecastDataPoint) (ret iface.Cond, err error) {
	codemap := map[string]iface.WeatherCode{
		"clear-day":           iface.CodeSunny,
		"clear-night":         iface.CodeSunny,
		"rain":                iface.CodeLightRain,
		"snow":                iface.CodeLightSnow,
		"sleet":               iface.CodeLightSleet,
		"wind":                iface.CodePartlyCloudy,
		"fog":                 iface.CodeFog,
		"cloudy":              iface.CodeCloudy,
		"partly-cloudy-day":   iface.CodePartlyCloudy,
		"partly-cloudy-night": iface.CodePartlyCloudy,
		"thunderstorm":        iface.CodeThunderyShowers,
	}

	if dp.Time == nil {
		return iface.Cond{}, fmt.Errorf("The forecast.io response did not provide a time for the weather condition")
	}
	ret.Time = time.Unix(int64(*dp.Time), 0)

	ret.Code = iface.CodeUnknown
	if val, ok := codemap[dp.Icon]; ok {
		ret.Code = val
	}
	ret.Desc = dp.Summary

	ret.TempC = dp.Temperature
	ret.FeelsLikeC = dp.ApparentTemperature

	if dp.PrecipProbability != nil {
		var p int = int(*dp.PrecipProbability * 100)
		ret.ChanceOfRainPercent = &p
	}

	if dp.PrecipIntensity != nil && *dp.PrecipIntensity >= 0 {
		var p float32 = *dp.PrecipIntensity / 1000
		ret.PrecipM = &p
	}

	if dp.Visibility != nil && *dp.Visibility >= 0 {
		var p float32 = *dp.Visibility * 1000
		ret.VisibleDistM = &p
	}

	if dp.WindSpeed != nil && *dp.WindSpeed >= 0 {
		ret.WindspeedKmph = dp.WindSpeed
	}

	//ret.WindGustKmph not provided by forecast.io :(

	if dp.WindBearing != nil && *dp.WindBearing >= 0 {
		var p int = int(*dp.WindBearing) % 360
		ret.WinddirDegree = &p
	}

	return ret, nil
}

func (c *forecastConfig) Setup() {
	flag.StringVar(&c.apiKey, "forecast-api-key", "", "forecast backend: the api `KEY` to use")
	flag.StringVar(&c.language, "forecast-lang", "en", "forecast backend: the `LANGUAGE` to request from forecast.io")
	flag.BoolVar(&c.debug, "forecast-debug", false, "forecast backend: print raw requests and responses")
}

func (c *forecastConfig) Fetch(location string, numdays int) iface.Data {
	var ret iface.Data

	if len(c.apiKey) == 0 {
		log.Fatal("No forecast.io API key specified.")
	}
	requri := fmt.Sprintf(forecastWuri, c.apiKey, location, c.language)

	if c.debug {
		log.Println("Weather request:", requri)
	}

	res, err := http.Get(requri)
	if err != nil {
		log.Fatal("Unable to get weather data: ", err)
	} else if res.StatusCode != 200 {
		log.Fatal("Unable to get weather data: http status ", res.StatusCode)
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	if c.debug {
		log.Println("Weather response:", string(body))
	}

	var resp forecastResponse
	if err = json.Unmarshal(body, &resp); err != nil {
		log.Println(err)
	}

	ret.Location = fmt.Sprintf("%f:%f", *resp.Latitude, *resp.Longitude)
	var reqLatLon iface.LatLon
	reqLatLon.Latitude = *resp.Latitude
	reqLatLon.Longitude = *resp.Longitude
	ret.GeoLoc = &reqLatLon

	if ret.Current, err = forecastParseCond(resp.Currently); err != nil {
		log.Fatalf("Could not parse current weather condition: %v", err)
	}
	ret.Forecast = forecastParseDaily(resp.Hourly, numdays)

	return ret
}

func init() {
	iface.AllBackends["forecast.io"] = &forecastConfig{}
}
