package backends

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/schachmat/wego/iface"
)

type weatherApiResponse struct {
	Location struct {
		Name    string `json:"name"`
		Country string `json:"country"`
	} `json:"location"`
	Forecast struct {
		List []forecastBlock `json:"forecastday"`
	} `json:"forecast"`
}

type forecastBlock struct {
	Date      time.Time `json:"date"`
	DateEpoch int64     `json:"date_epoch"`
	Day       struct {
		TempC        float32 `json:"avgtemp_c"`
		Humidity     int     `json:"avghumidity"`
		MaxWindSpeed float32 `json:"maxwind_kph"`
		Weather      struct {
			Description string `json:"text"`
			Code        int    `json:"code"`
		} `json:"condition"`
		Hour []hourlyWeather `json:"hour"`
	} `json:"day"`
}

type hourlyWeather struct {
	TimeEpoch  int64   `json:"time_epoch"`
	TempC      float32 `json:"temp_c"`
	FeelsLikeC float32 `json:"feelslike_c"`
	Humidity   int     `json:"humidity"`
	Condition  struct {
		Code int    `json:"code"`
		Desc string `json:"text"`
	} `json:"condition"`
	WindspeedKmph       *float32 `json:"wind_kph"`
	WinddirDegree       int      `json:"wind_degree"`
	ChanceOfRainPercent int      `json:"chance_of_rain"`
}

type weatherApiConfig struct {
	apiKey string
	debug  bool
}

const (
	weatherApiURI = "http://api.weatherapi.com/v1/forecast.json?key=%s&q=%s&days=3&aqi=no&alerts=no"
)

var (
	codemapping = map[int]iface.WeatherCode{
		1000: iface.CodeSunny,
		1003: iface.CodePartlyCloudy,
		1006: iface.CodeCloudy,
		1009: iface.CodeVeryCloudy,
		1030: iface.CodeVeryCloudy,
		1063: iface.CodeLightRain,
		1066: iface.CodeLightSnowShowers,
		1069: iface.CodeLightSnowShowers,
		1071: iface.CodeLightShowers,
		1087: iface.CodeThunderyShowers,
		1114: iface.CodeHeavySnow,
		1117: iface.CodeHeavySnow,
		1135: iface.CodeFog,
		1147: iface.CodeFog,
		1150: iface.CodeLightRain,
		1153: iface.CodeLightRain,
		1168: iface.CodeLightRain,
		1171: iface.CodeHeavyRain,
		1180: iface.CodeLightRain,
		1183: iface.CodeLightRain,
		1186: iface.CodeHeavyRain,
		1189: iface.CodeHeavyRain,
		1192: iface.CodeHeavyRain,
		1195: iface.CodeHeavyRain,
		1198: iface.CodeLightRain,
		1201: iface.CodeHeavyRain,
		1204: iface.CodeLightSleet,
		1207: iface.CodeLightSleetShowers,
		1210: iface.CodeLightSnow,
		1213: iface.CodeLightSnow,
		1216: iface.CodeHeavySnow,
		1219: iface.CodeHeavySnow,
		1222: iface.CodeHeavySnow,
		1225: iface.CodeHeavySnow,
		1237: iface.CodeHeavySnow,
		1240: iface.CodeLightShowers,
		1243: iface.CodeHeavyShowers,
		1246: iface.CodeThunderyShowers,
		1249: iface.CodeLightSleetShowers,
		1252: iface.CodeLightSleetShowers,
		1255: iface.CodeLightSnowShowers,
		1258: iface.CodeHeavySnowShowers,
		1261: iface.CodeLightSnowShowers,
		1264: iface.CodeHeavySnowShowers,
		1273: iface.CodeThunderyShowers,
		1276: iface.CodeThunderyHeavyRain,
		1279: iface.CodeThunderySnowShowers,
		1282: iface.CodeThunderySnowShowers,
	}
)

func (c *weatherApiConfig) Setup() {
	flag.StringVar(&c.apiKey, "wth-api-key", "", "weatherapi backend: the api `Key` to use")
	flag.BoolVar(&c.debug, "wth-debug", false, "weatherapi backend: print raw requests and responses")
}

func (c *weatherApiConfig) fetch(url string) (*weatherApiResponse, error) {
	res, err := http.Get(url)
	if c.debug {
		fmt.Printf("Fetching %s\n", url)
	}
	if err != nil {
		return nil, fmt.Errorf("Unable to get (%s) %v", url, err)
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("Unable to read response body (%s): %v", url, err)
	}

	if c.debug {
		fmt.Printf("Response (%s):\n%s\n", url, string(body))
	}

	var resp weatherApiResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("Unable to unmarshal response (%s): %v\nThe json body is: %s", url, err, string(body))
	}

	return &resp, nil
}

func (c *weatherApiConfig) parseDaily(dataBlock []forecastBlock, numdays int) []iface.Day {
	var forecast []iface.Day
	var day *iface.Day

	for _, dayData := range dataBlock {
		for _, data := range dayData.Day.Hour {
			slot, err := c.parseCond(data)
			if err != nil {
				log.Println("Error parsing weather condition:", err)
				continue
			}
			if day == nil {
				day = new(iface.Day)
				day.Date = slot.Time
			}
			if day.Date.Day() == slot.Time.Day() {
				day.Slots = append(day.Slots, slot)
			}
			if day.Date.Day() != slot.Time.Day() {
				forecast = append(forecast, *day)
				if len(forecast) >= numdays {
					break
				}
				day = new(iface.Day)
				day.Date = slot.Time
				day.Slots = append(day.Slots, slot)
			}
		}
	}

	return forecast
}

func (c *weatherApiConfig) parseCond(forecastInfo hourlyWeather) (iface.Cond, error) {
	var ret iface.Cond

	ret.Code = iface.CodeUnknown
	ret.Desc = forecastInfo.Condition.Desc
	ret.Humidity = &(forecastInfo.Humidity)
	ret.TempC = &(forecastInfo.TempC)
	ret.FeelsLikeC = &(forecastInfo.FeelsLikeC)
	ret.WindspeedKmph = forecastInfo.WindspeedKmph
	ret.WinddirDegree = &forecastInfo.WinddirDegree
	ret.ChanceOfRainPercent = &forecastInfo.ChanceOfRainPercent

	if val, ok := codemapping[forecastInfo.Condition.Code]; ok {
		ret.Code = val
	}

	ret.Time = time.Unix(forecastInfo.TimeEpoch, 0)

	return ret, nil
}

func (c *weatherApiConfig) parseCurCond(forecastInfo forecastBlock) (iface.Cond, error) {
	var ret iface.Cond

	ret.Code = iface.CodeUnknown
	ret.Desc = forecastInfo.Day.Weather.Description
	ret.Humidity = &(forecastInfo.Day.Humidity)
	ret.TempC = &(forecastInfo.Day.TempC)

	if val, ok := codemapping[forecastInfo.Day.Weather.Code]; ok {
		ret.Code = val
	}

	ret.Time = time.Unix(forecastInfo.DateEpoch, 0)

	return ret, nil
}

func (c *weatherApiConfig) Fetch(location string, numdays int) iface.Data {
	var ret iface.Data

	if len(c.apiKey) == 0 {
		log.Fatal("No openweathermap.org API key specified.\nYou have to register for one at https://home.openweathermap.org/users/sign_up")
	}

	resp, err := c.fetch(fmt.Sprintf(weatherApiURI, c.apiKey, location))
	if err != nil {
		log.Fatalf("Failed to fetch weather data: %v\n", err)
	}
	fmt.Println(resp)
	ret.Current, err = c.parseCurCond(resp.Forecast.List[0])
	ret.Location = fmt.Sprintf("%s, %s", resp.Location.Name, resp.Location.Country)

	if err != nil {
		log.Fatalf("Failed to fetch weather data: %v\n", err)
	}

	if numdays == 0 {
		return ret
	}
	ret.Forecast = c.parseDaily(resp.Forecast.List, numdays)

	return ret
}

func init() {
	iface.AllBackends["weatherapi"] = &weatherApiConfig{}
}
