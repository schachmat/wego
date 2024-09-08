package backends

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/schachmat/wego/iface"
)

type openmeteoConfig struct {
	apiKey   string
	language string
	debug    bool
}

type curCond struct {
	Time                int64    `json:"time"`
	Interval            int      `json:"interval"`
	Temperature2M       *float32 `json:"temperature_2m"`
	ApparentTemperature *float32 `json:"apparent_temperature"`
	IsDay               int      `json:"is_day"`
	WeatherCode         int      `json:"weather_code"`
	WindDirection10M    *int     `json:"wind_direction_10m"`
}

type Daily struct {
	Time                   []int64    `json:"time"`
	WeatherCode            []int      `json:"weather_code"`
	Temperature2MMax       []*float32 `json:"temperature_2m_max"`
	ApparentTemperatureMax []*float32 `json:"apparent_temperature_max"`
	Sunrise                []int64    `json:"sunrise"`
	Sunset                 []int64    `json:"sunset"`
}
type HourlyUnits struct {
	Time                string `json:"time"`
	Temperature2M       string `json:"temperature_2m"`
	ApparentTemperature string `json:"apparent_temperature"`
	WeatherCode         string `json:"weather_code"`
}
type Hourly struct {
	Time                []int64    `json:"time"`
	Temperature2M       []*float32 `json:"temperature_2m"`
	ApparentTemperature []*float32 `json:"apparent_temperature"`
	WeatherCode         []int      `json:"weather_code"`
	WindDirection10M    []*int     `json:"wind_direction_10m"`
}

type openmeteoResponse struct {
	Latitude             float64 `json:"latitude"`
	Longitude            float64 `json:"longitude"`
	GenerationtimeMs     float64 `json:"generationtime_ms"`
	UtcOffsetSeconds     int     `json:"utc_offset_seconds"`
	Timezone             string  `json:"timezone"`
	TimezoneAbbreviation string  `json:"timezone_abbreviation"`
	Elevation            float64 `json:"elevation"`
	CurrentUnits         struct {
		Time                string `json:"time"`
		Interval            string `json:"interval"`
		Temperature2M       string `json:"temperature_2m"`
		ApparentTemperature string `json:"apparent_temperature"`
		IsDay               string `json:"is_day"`
		WeatherCode         string `json:"weather_code"`
	} `json:"current_units"`
	Current     curCond     `json:"current"`
	HourlyUnits HourlyUnits `json:"hourly_units"`
	Hourly      Hourly      `json:"hourly"`
	DailyUnits  struct {
		Time                   string `json:"time"`
		WeatherCode            string `json:"weather_code"`
		Temperature2MMax       string `json:"temperature_2m_max"`
		ApparentTemperatureMax string `json:"apparent_temperature_max"`
		Sunrise                string `json:"sunrise"`
		Sunset                 string `json:"sunset"`
	} `json:"daily_units"`
	Daily Daily
}

const (
	openmeteoURI = "https://api.open-meteo.com/v1/forecast?"
)

var (
	codemap = map[int]iface.WeatherCode{
		0:  iface.CodeSunny,
		1:  iface.CodePartlyCloudy,
		2:  iface.CodePartlyCloudy,
		3:  iface.CodePartlyCloudy,
		45: iface.CodeFog,
		48: iface.CodeFog,
		51: iface.CodeLightRain,
		53: iface.CodeLightRain,
		55: iface.CodeLightRain,
		56: iface.CodeLightSleet,
		57: iface.CodeLightSleet,
		61: iface.CodeLightShowers,
		63: iface.CodeLightShowers,
		65: iface.CodeLightShowers,
		66: iface.CodeHeavyRain,
		67: iface.CodeHeavyRain,
	}
)

func (opmeteo *openmeteoConfig) Setup() {
	flag.StringVar(&opmeteo.apiKey, "openmeteo-api-key", "", "openmeteo backend: the api `KEY` to use if commercial usage")
	flag.BoolVar(&opmeteo.debug, "openmeteo-debug", false, "openmeteo backend: print raw requests and responses")
}

func (opmeteo *openmeteoConfig) parseDaily(dailyInfo Hourly) []iface.Day {
	var forecast []iface.Day
	var day *iface.Day

	for ind, dayTime := range dailyInfo.Time {

		cond := new(iface.Cond)

		cond.Code = codemap[dailyInfo.WeatherCode[ind]]
		cond.TempC = dailyInfo.Temperature2M[ind]
		cond.FeelsLikeC = dailyInfo.ApparentTemperature[ind]
		cond.Time = time.Unix(dayTime, 0)
		cond.WinddirDegree = dailyInfo.WindDirection10M[ind]

		if day == nil {
			day = new(iface.Day)
			day.Date = cond.Time
		}
		if day.Date.Day() == cond.Time.Day() {
			day.Slots = append(day.Slots, *cond)
		}
		if day.Date.Day() != cond.Time.Day() {
			forecast = append(forecast, *day)

			day = new(iface.Day)
			day.Date = cond.Time
			day.Slots = append(day.Slots, *cond)
		}
	}

	return forecast
}

func parseCurCond(current curCond) (ret iface.Cond) {

	ret.Time = time.Unix(current.Time, 0)

	ret.Code = iface.CodeUnknown
	if val, ok := codemap[current.WeatherCode]; ok {
		ret.Code = val
	}

	ret.TempC = current.Temperature2M
	ret.FeelsLikeC = current.ApparentTemperature
	ret.WinddirDegree = current.WindDirection10M
	return ret

}

func (opmeteo *openmeteoConfig) Fetch(location string, numdays int) iface.Data {
	var ret iface.Data
	var params []string
	var loc string

	if matched, err := regexp.MatchString(`^-?[0-9]*(\.[0-9]+)?,-?[0-9]*(\.[0-9]+)?$`, location); matched && err == nil {
		s := strings.Split(location, ",")
		loc = fmt.Sprintf("latitude=%s&longitude=%s", s[0], s[1])
	}
	if len(location) > 0 {
		params = append(params, loc)
	}
	params = append(params, "current=temperature_2m,apparent_temperature,is_day,weather_code&hourly=temperature_2m,apparent_temperature,weather_code,wind_direction_10m&daily=weather_code,temperature_2m_max,apparent_temperature_max,sunrise,sunset&timeformat=unixtime&forecast_days=3")

	requri := openmeteoURI + strings.Join(params, "&")

	res, err := http.Get(requri)
	if err != nil {
		log.Fatal("Unable to get weather data: ", err)
	} else if res.StatusCode != 200 {
		log.Fatal("Unable to get weather data: http status ", res.StatusCode)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	if opmeteo.debug {
		log.Println("Weather request:", requri)
		b, _ := json.MarshalIndent(body, "", "\t")
		fmt.Println("Weather response:", string(b))
	}

	var resp openmeteoResponse
	if err = json.Unmarshal(body, &resp); err != nil {
		log.Println(err)
	}

	ret.Current = parseCurCond(resp.Current)
	ret.Location = location

	forecast := opmeteo.parseDaily(resp.Hourly)

	if len(forecast) > 0 {
		forecast[0].Astronomy.Sunset = time.Unix(resp.Daily.Sunset[0], 0)
		forecast[0].Astronomy.Sunrise = time.Unix(resp.Daily.Sunrise[0], 0)
		ret.Forecast = forecast
	}

	return ret
}

func init() {
	iface.AllBackends["openmeteo"] = &openmeteoConfig{}
}
