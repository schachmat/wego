package backends

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"time"

	"github.com/schachmat/wego/iface"
)

type forecastConfig struct {
	apiKey string
	lang   string
	debug  bool
	tz     *time.Location
}

type forecastDataPoint struct {
	Time                *int64   `json:"time"`
	Summary             string   `json:"summary"`
	Icon                string   `json:"icon"`
	SunriseTime         *int64   `json:"sunriseTime"`
	SunsetTime          *int64   `json:"sunsetTime"`
	PrecipIntensity     *float32 `json:"precipIntensity"`
	PrecipProb          *float32 `json:"precipProbability"`
	Temperature         *float32 `json:"temperature"`
	ApparentTemperature *float32 `json:"apparentTemperature"`
	WindSpeed           *float32 `json:"windSpeed"`
	WindBearing         *float32 `json:"windBearing"`
	Visibility          *float32 `json:"visibility"`
	Humidity            *float32 `json:"humidity"`
}

type forecastDataBlock struct {
	Summary string              `json:"summary"`
	Icon    string              `json:"icon"`
	Data    []forecastDataPoint `json:"data"`
}

type forecastResponse struct {
	Latitude  *float32          `json:"latitude"`
	Longitude *float32          `json:"longitude"`
	Timezone  *string           `json:"timezone"`
	Currently forecastDataPoint `json:"currently"`
	Hourly    forecastDataBlock `json:"hourly"`
	Daily     forecastDataBlock `json:"daily"`
}

const (
	// see https://darksky.net/dev/docs
	// see also https://github.com/mlbright/forecast
	//https://api.darksky.net/forecast/APIKEY/LATITUDE,LONGITUDE
	forecastWuri = "https://api.darksky.net/forecast/%s/%s?units=ca&lang=%s&exclude=minutely,alerts,flags&extend=hourly"
)

func (c *forecastConfig) parseAstro(cur *iface.Day, days []forecastDataPoint) {
	for _, day := range days {
		if day.Time != nil && cur.Date.Day() == time.Unix(*day.Time, 0).In(c.tz).Day() {
			if day.SunriseTime != nil {
				cur.Astronomy.Sunrise = time.Unix(*day.SunriseTime, 0).In(c.tz)
			}
			if day.SunsetTime != nil {
				cur.Astronomy.Sunset = time.Unix(*day.SunsetTime, 0).In(c.tz)
			}
			return
		}
	}
}

func (c *forecastConfig) parseDaily(hours, days forecastDataBlock, numdays int) []iface.Day {
	var forecast []iface.Day
	var day *iface.Day

	for _, hourData := range hours.Data {
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
			c.parseAstro(day, days.Data)
		}

		day.Slots = append(day.Slots, slot)
	}
	return append(forecast, *day)
}

func (c *forecastConfig) parseCond(dp forecastDataPoint) (ret iface.Cond, err error) {
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
	ret.Time = time.Unix(*dp.Time, 0).In(c.tz)

	ret.Code = iface.CodeUnknown
	if val, ok := codemap[dp.Icon]; ok {
		ret.Code = val
	}
	ret.Desc = dp.Summary

	ret.TempC = dp.Temperature
	ret.FeelsLikeC = dp.ApparentTemperature

	if dp.PrecipProb != nil && *dp.PrecipProb >= 0 && *dp.PrecipProb <= 1 {
		p := int(*dp.PrecipProb * 100)
		ret.ChanceOfRainPercent = &p
	}

	if dp.PrecipIntensity != nil && *dp.PrecipIntensity >= 0 {
		p := *dp.PrecipIntensity / 1000
		ret.PrecipM = &p
	}

	if dp.Visibility != nil && *dp.Visibility >= 0 {
		p := *dp.Visibility * 1000
		ret.VisibleDistM = &p
	}

	if dp.WindSpeed != nil && *dp.WindSpeed >= 0 {
		ret.WindspeedKmph = dp.WindSpeed
	}

	//ret.WindGustKmph not provided by darksky.net :(

	if dp.WindBearing != nil && *dp.WindBearing >= 0 {
		p := int(*dp.WindBearing) % 360
		ret.WinddirDegree = &p
	}

	if dp.Humidity != nil && *dp.Humidity >= 0 && *dp.Humidity <= 1 {
		p := int(*dp.Humidity * 100)
		ret.Humidity = &p
	}

	return ret, nil
}

func (c *forecastConfig) fetch(url string) (*forecastResponse, error) {
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

	var resp forecastResponse
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

func (c *forecastConfig) fetchToday(location string) ([]iface.Cond, error) {
	location = fmt.Sprintf("%s,%d", location, time.Now().Unix())

	resp, err := c.fetch(fmt.Sprintf(forecastWuri, c.apiKey, location, c.lang))
	if err != nil {
		return nil, fmt.Errorf("Failed to fetch todays weather data: %v\n", err)
	}

	days := c.parseDaily(resp.Hourly, resp.Daily, 1)
	if len(days) < 1 {
		return nil, fmt.Errorf("Failed to parse today\n")
	}
	return days[0].Slots, nil
}

func (c *forecastConfig) Setup() {
	flag.StringVar(&c.apiKey, "forecast-api-key", "", "forecast backend: the api `KEY` to use")
	flag.StringVar(&c.lang, "forecast-lang", "en", "forecast backend: the `LANGUAGE` to request from forecast.io")
	flag.BoolVar(&c.debug, "forecast-debug", false, "forecast backend: print raw requests and responses")
}

func (c *forecastConfig) Fetch(location string, numdays int) iface.Data {
	var ret iface.Data
	todayChan := make(chan []iface.Cond)

	if len(c.apiKey) == 0 {
		log.Fatal("No forecast.io API key specified.\nYou have to register for one at https://developer.forecast.io/register")
	}
	if matched, err := regexp.MatchString(`^-?[0-9]*(\.[0-9]+)?,-?[0-9]*(\.[0-9]+)?$`, location); !matched || err != nil {
		log.Fatalf("Error: The forecast.io backend only supports latitude,longitude pairs as location.\nInstead of `%s` try `40.748,-73.985` for example to get a forecast for New York", location)
	}

	c.tz = time.Local

	go func() {
		slots, err := c.fetchToday(location)
		if err != nil {
			log.Fatalf("Failed to fetch todays weather data: %v\n", err)
		}
		todayChan <- slots
	}()

	resp, err := c.fetch(fmt.Sprintf(forecastWuri, c.apiKey, location, c.lang))
	if err != nil {
		log.Fatalf("Failed to fetch weather data: %v\n", err)
	}

	if resp.Latitude == nil || resp.Longitude == nil {
		log.Println("nil response for latitude,longitude")
		ret.Location = location
	} else {
		ret.GeoLoc = &iface.LatLon{Latitude: *resp.Latitude, Longitude: *resp.Longitude}
		ret.Location = fmt.Sprintf("%f,%f", *resp.Latitude, *resp.Longitude)
	}

	if ret.Current, err = c.parseCond(resp.Currently); err != nil {
		log.Fatalf("Could not parse current weather condition: %v", err)
	}

	if numdays >= 1 {
		ret.Forecast = c.parseDaily(resp.Hourly, resp.Daily, numdays)

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
	iface.AllBackends["forecast.io"] = &forecastConfig{}
}
