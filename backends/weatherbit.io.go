package backends

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/schachmat/wego/iface"
)

type weatherbitConfig struct {
	apiKey string
	lang   string
	debug  bool
	tz     *time.Location
}

type weatherbitDataPoint struct {
	Time                *string            `json:"timestamp_local"`
	Weather             *weatherbitWeather `json:"weather"`
	PrecipIntensity     *float32           `json:"precip"`
	PrecipProb          *float32           `json:"pop"`
	Temperature         *float32           `json:"temp"`
	ApparentTemperature *float32           `json:"app_temp"`
	WindSpeed           *float32           `json:"wind_spd"`
	WindGustSpeed       *float32           `json:"wind_gust_spd"`
	WindBearing         *float32           `json:"wind_dir"`
	Humidity            *int32             `json:"rh"`
	Visibility          *float32           `json:"vis"`
}

type weatherbitWeather struct {
	Summary *string `json:"description"`
	Code    int32   `json:"code"`
	Icon    *string `json:"icon"`
}

type weatherbitResponse struct {
	Latitude  *float32              `json:"lat,string"`
	Longitude *float32              `json:"lon,string"`
	CityName  *string               `json:"city_name"`
	Timezone  *string               `json:"timezone"`
	Data      []weatherbitDataPoint `json:"data"`
}

const (
	weatherbitUriLatLon = "https://api.weatherbit.io/v2.0/forecast/3hourly?units=M&lang=%s&key=%s&lat=%s&lon=%s&days=%d"
	weatherbitUriName   = "https://api.weatherbit.io/v2.0/forecast/3hourly?units=M&lang=%s&key=%s&city=%s&days=%d"
)

func (c *weatherbitConfig) parseDaily(dataInfo []weatherbitDataPoint, numdays int) []iface.Day {
	var forecast []iface.Day
	var day *iface.Day

	for _, hourData := range dataInfo {
		slot, err := c.parseCond(hourData)
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
		}

		day.Slots = append(day.Slots, slot)
	}
	return append(forecast, *day)
}

func (c *weatherbitConfig) parseCond(dp weatherbitDataPoint) (ret iface.Cond, err error) {
	codemap := map[string]iface.WeatherCode{
		"t01d": iface.CodeThunderyShowers,
		"t01n": iface.CodeThunderyShowers,
		"t02d": iface.CodeThunderyShowers,
		"t02n": iface.CodeThunderyShowers,
		"t03d": iface.CodeThunderyHeavyRain,
		"t03n": iface.CodeThunderyHeavyRain,
		"t04d": iface.CodeThunderyShowers,
		"t04n": iface.CodeThunderyShowers,
		"t05d": iface.CodeThunderySnowShowers,
		"t05n": iface.CodeThunderySnowShowers,

		"d01d": iface.CodeLightShowers,
		"d01n": iface.CodeLightShowers,
		"d02d": iface.CodeLightShowers,
		"d02n": iface.CodeLightShowers,
		"d03d": iface.CodeLightShowers,
		"d03n": iface.CodeLightShowers,

		"r01d": iface.CodeLightRain,
		"r01n": iface.CodeLightRain,
		"r02d": iface.CodeLightRain,
		"r02n": iface.CodeLightRain,
		"r03d": iface.CodeHeavyRain,
		"r03n": iface.CodeHeavyRain,
		"r04d": iface.CodeHeavyRain,
		"r04n": iface.CodeHeavyRain,
		"r05d": iface.CodeLightShowers,
		"r05n": iface.CodeLightShowers,
		"r06d": iface.CodeHeavyShowers,
		"r06n": iface.CodeHeavyShowers,
		"u00d": iface.CodeHeavyShowers,
		"u00n": iface.CodeHeavyShowers,

		"s01d": iface.CodeLightSnow,
		"s01n": iface.CodeLightSnow,
		"s02d": iface.CodeLightSnow,
		"s02n": iface.CodeLightSnow,
		"s03d": iface.CodeHeavySnow,
		"s03n": iface.CodeHeavySnow,
		"s04d": iface.CodeHeavySnow,
		"s04n": iface.CodeHeavySnow,
		"s05d": iface.CodeHeavySnow,
		"s05n": iface.CodeHeavySnow,

		"a01d": iface.CodeFog,
		"a01n": iface.CodeFog,
		"a02d": iface.CodeFog,
		"a02n": iface.CodeFog,
		"a03d": iface.CodeFog,
		"a03n": iface.CodeFog,
		"a04d": iface.CodeFog,
		"a04n": iface.CodeFog,
		"a05d": iface.CodeFog,
		"a05n": iface.CodeFog,
		"a06d": iface.CodeFog,
		"a06n": iface.CodeFog,

		"c01d": iface.CodeSunny,
		"c01n": iface.CodeSunny,
		"c02d": iface.CodePartlyCloudy,
		"c02n": iface.CodePartlyCloudy,
		"c03d": iface.CodePartlyCloudy,
		"c03n": iface.CodePartlyCloudy,
		"c04d": iface.CodeVeryCloudy,
		"c04n": iface.CodeVeryCloudy,
	}

	if dp.Time == nil {
		return iface.Cond{}, fmt.Errorf("The weatherbit.io response did not provide a time for the weather condition")
	}

	cDate, _ := time.Parse("2006-01-02T15:04:05", *dp.Time)
	ret.Time = cDate.In(c.tz)

	ret.Code = iface.CodeUnknown
	if dp.Weather.Icon != nil {
		if val, ok := codemap[*dp.Weather.Icon]; ok {
			ret.Code = val
		}
	}

	if dp.Weather.Summary != nil {
		ret.Desc = *dp.Weather.Summary
	}

	ret.TempC = dp.Temperature
	ret.FeelsLikeC = dp.ApparentTemperature

	if dp.PrecipProb != nil && *dp.PrecipProb >= 0 && *dp.PrecipProb <= 100 {
		p := int(*dp.PrecipProb)
		ret.ChanceOfRainPercent = &p
	}

	if dp.PrecipIntensity != nil && *dp.PrecipIntensity >= 0 {
		p := *dp.PrecipIntensity / 1000
		ret.PrecipM = &p
	}

	if dp.WindSpeed != nil && *dp.WindSpeed >= 0 {
		windKmph := *dp.WindSpeed * 3.6
		ret.WindspeedKmph = &windKmph
	}

	if dp.WindGustSpeed != nil && *dp.WindGustSpeed >= 0 {
		gustKmph := *dp.WindSpeed * 3.6
		ret.WindGustKmph = &gustKmph
	}

	if dp.WindBearing != nil && *dp.WindBearing >= 0 {
		p := int(*dp.WindBearing) % 360
		ret.WinddirDegree = &p
	}

	if dp.Humidity != nil && *dp.Humidity >= 0 && *dp.Humidity <= 100 {
		p := int(*dp.Humidity)
		ret.Humidity = &p
	}

	if dp.Visibility != nil && *dp.Visibility >= 0 {
		p := *dp.Visibility * 1000
		ret.VisibleDistM = &p
	}

	return ret, nil
}

func (c *weatherbitConfig) fetch(url string) (*weatherbitResponse, error) {
	res, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("Unable to get (%s): %v", url, err)
	} else if res.StatusCode != 200 {
		return nil, fmt.Errorf("Unable to get (%s): http status %d", url, res.StatusCode)
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("Unable to read response body (%s): %v", url, err)
	}

	if c.debug {
		log.Printf("Response (%s): %s\n", url, string(body))
	}

	var resp weatherbitResponse
	if err = json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("Unable to unmarshal response (%s): %v\nThe json body is: %s", url, err, string(body))
	}

	if resp.Timezone == nil {
		log.Printf("No timezone set in response (%s)", url)
	} else {
		c.tz, err = time.LoadLocation(*resp.Timezone)
		if err != nil {
			log.Printf("Unknown Timezone used in response (%s)", url)
		}
	}
	return &resp, nil
}

func (c *weatherbitConfig) fetchToday(location string) ([]iface.Cond, error) {
	matched, err := regexp.MatchString(`^-?[0-9]*(\.[0-9]+)?,-?[0-9]*(\.[0-9]+)?$`, location)
	if err != nil {
		log.Fatalf("Error: Unable to parse location '%s'", location)
	}

	var resp *weatherbitResponse
	if matched {
		locationParts := strings.Split(location, ",")
		resp, err = c.fetch(fmt.Sprintf(weatherbitUriLatLon, c.lang, c.apiKey, locationParts[0], locationParts[1], 1))
	} else {
		resp, err = c.fetch(fmt.Sprintf(weatherbitUriName, c.lang, c.apiKey, url.QueryEscape(location), 1))
	}
	if err != nil {
		return nil, fmt.Errorf("Failed to fetch todays weather data: %v\n", err)
	}

	days := c.parseDaily(resp.Data, 1)
	if len(days) < 1 {
		return nil, fmt.Errorf("Failed to parse today\n")
	}
	return days[0].Slots, nil
}

func (c *weatherbitConfig) Setup() {
	flag.StringVar(&c.apiKey, "weatherbit-api-key", "", "weatherbit backend: the api `KEY` to use")
	flag.StringVar(&c.lang, "weatherbit-lang", "en", "weatherbit backend: the `LANGUAGE` to request from weatherbit.io")
	flag.BoolVar(&c.debug, "weatherbit-debug", false, "weatherbit backend: print raw requests and responses")
}

func (c *weatherbitConfig) Fetch(location string, numdays int) iface.Data {
	var ret iface.Data
	todayChan := make(chan []iface.Cond)

	if len(c.apiKey) == 0 {
		log.Fatal("No weatherbit.io API key specified.\nYou have to register for one at https://www.weatherbit.io/account/create")
	}

	c.tz = time.Local

	go func() {
		slots, err := c.fetchToday(location)
		if err != nil {
			log.Fatalf("Failed to fetch todays weather data: %v\n", err)
		}
		todayChan <- slots
	}()

	matched, err := regexp.MatchString(`^-?[0-9]*(\.[0-9]+)?,-?[0-9]*(\.[0-9]+)?$`, location)
	if err != nil {
		log.Fatalf("Error: Unable to parse location '%s'", location)
	}

	var resp *weatherbitResponse
	if matched {
		locationParts := strings.Split(location, ",")
		resp, err = c.fetch(fmt.Sprintf(weatherbitUriLatLon, c.lang, c.apiKey, locationParts[0], locationParts[1], numdays))
	} else {
		resp, err = c.fetch(fmt.Sprintf(weatherbitUriName, c.lang, c.apiKey, url.QueryEscape(location), numdays))
	}

	if err != nil {
		log.Fatalf("Failed to fetch weather data: %v\n", err)
	}

	if resp.CityName != nil {
		ret.Location = fmt.Sprintf("%s (%f,%f)", *resp.CityName, *resp.Latitude, *resp.Longitude)
	} else if resp.Latitude == nil || resp.Longitude == nil {
		log.Println("nil response for latitude,longitude")
		ret.Location = location
	} else {
		ret.GeoLoc = &iface.LatLon{Latitude: *resp.Latitude, Longitude: *resp.Longitude}
		ret.Location = fmt.Sprintf("%f,%f", *resp.Latitude, *resp.Longitude)
	}

	if ret.Current, err = c.parseCond(resp.Data[0]); err != nil {
		log.Fatalf("Could not parse current weather condition: %v", err)
	}

	if numdays >= 1 {
		ret.Forecast = c.parseDaily(resp.Data, numdays)

		var tHistory, tFuture = <-todayChan, ret.Forecast[0].Slots
		var tRet []iface.Cond
		h, f := 0, 0

		// merge forecast and history from current day
		for h < len(tHistory) || f < len(tFuture) {
			if f >= len(tFuture) {
				tRet = append(tRet, tHistory[h])
				h++
			} else if h >= len(tHistory) || tHistory[h].Time.After(tFuture[f].Time) {
				tRet = append(tRet, tFuture[f])
				f++
			} else if tHistory[h].Time.Before(tFuture[f].Time) {
				tRet = append(tRet, tHistory[h])
				h++
			} else {
				tRet = append(tRet, tFuture[f])
				h++
				f++
			}
		}
		ret.Forecast[0].Slots = tRet
	}
	return ret
}

func init() {
	iface.AllBackends["weatherbit.io"] = &weatherbitConfig{}
}
