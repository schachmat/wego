package backends

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/schachmat/wego/iface"
)

type forecastConfig struct {
	apiKey    string
	latitude  string
	longitude string
	debug     bool
}

type forecastDataPoint struct {
	Time                   float64  `json:"time"`
	Summary                string   `json:"summary"`
	Icon                   string   `json:"icon"`
	SunriseTime            float32  `json:"sunriseTime"`
	SunsetTime             float32  `json:"sunsetTime"`
	PrecipIntensity        *float32 `json:"precipIntensity"`
	PrecipIntensityMax     *float32 `json:"precipIntensityMax"`
	PrecipIntensityMaxTime float32  `json:"precipIntensityMaxTime"`
	PrecipProbability      float32  `json:"precipProbability"`
	PrecipType             string   `json:"precipType"`
	PrecipAccumulation     *float32 `json:"precipAccumulation"`
	Temperature            *float32 `json:"temperature"`
	TemperatureMin         *float32 `json:"temperatureMin"`
	TemperatureMinTime     float32  `json:"temperatureMinTime"`
	TemperatureMax         *float32 `json:"temperatureMax"`
	TemperatureMaxTime     float32  `json:"temperatureMaxTime"`
	ApparentTemperature    *float32 `json:"apparentTemperature"`
	DewPoint               *float32 `json:"dewPoint"`
	WindSpeed              *float32 `json:"windSpeed"`
	WindBearing            float32  `json:"windBearing"`
	CloudCover             *float32 `json:"cloudCover"`
	Humidity               *float32 `json:"humidity"`
	Pressure               *float32 `json:"pressure"`
	Visibility             *float32 `json:"visibility"`
	Ozone                  *float32 `json:"ozone"`
	MoonPhase              *float32 `json:"moonPhase"`
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
	forecastWuri = "https://api.forecast.io/forecast/%s/%s?units=si&lang=de&exclude=minutely,daily,alerts,flags&extend=hourly"
)

func (dp forecastDataPoint) Render() {
	var b []byte
	var err error
	b, err = json.MarshalIndent(dp, "", "\t")
	if err != nil {
		log.Fatal(err)
	}
	os.Stdout.Write(b)
}

func (db *forecastDataBlock) Convert(c *forecastConfig) []iface.Day {
	var forecast []iface.Day

	var day *iface.Day
	day = new(iface.Day)

	for cnt, dp := range db.Data {
		var slot iface.Cond
		slot = dp.Convert(c)

		//skip today
		// if slot.Time.Day() == time.Now().Day() {
		// 	continue
		// }

		// dp.Render()

		if c.debug {
			log.Printf("DataPoint: %02d\t%v\n", cnt, slot.Time)
		}

		if day == nil || day.Date.Day() != slot.Time.Day() {
			//is day already set?
			if len(day.Slots) >= 1 {
				if dp.TemperatureMax != nil && *dp.TemperatureMax >= 0 {
					day.MaxtempC = new(float32)
					*day.MaxtempC = *dp.TemperatureMax
				}

				if dp.TemperatureMax != nil && *dp.TemperatureMax >= 0 {
					day.MintempC = new(float32)
					*day.MintempC = *dp.TemperatureMin
				}

				forecast = append(forecast, *day)

				if c.debug {
					log.Printf("New Day: %02d\t%v\n", cnt, day)
					for i, cond := range day.Slots {
						log.Printf("New Day Slot: %02d\t%v\n", i, cond)
					}
				}
			}

			day = new(iface.Day)
			day.Date = slot.Time
			day.Slots = []iface.Cond{slot}
			// only add relevant Slots
		} else {
			if slot.Time.Hour() == 8 ||
				slot.Time.Hour() == 12 ||
				slot.Time.Hour() == 19 ||
				slot.Time.Hour() == 23 {
				day.Date = slot.Time
				day.Slots = append(day.Slots, slot)

				if c.debug {
					// log.Printf("Adding Slot: %02d\t>%p<\t>%v<\t>%v<\n", len(day.Slots), &slot, slot.Time, day)
				}
			} else if false {
				day.Date = slot.Time
				day.Slots = append(day.Slots, slot)
			}
		}
	}
	forecast = append(forecast, *day)

	return forecast
}

func (dp *forecastDataPoint) Convert(c *forecastConfig) iface.Cond {
	codemap := map[string]iface.WeatherCode{
		"wind":                iface.CodeUnknown,
		"hail":                iface.CodeUnknown,
		"tornado":             iface.CodeUnknown,
		"cloudy":              iface.CodeCloudy,
		"fog":                 iface.CodeFog,
		"rain":                iface.CodeLightRain,
		"sleet":               iface.CodeLightSleet,
		"snow":                iface.CodeLightSnow,
		"partly-cloudy-day":   iface.CodePartlyCloudy,
		"partly-cloudy-night": iface.CodePartlyCloudy,
		"clear-day":           iface.CodeSunny,
		"clear-night":         iface.CodeSunny,
		"thunderstorm":        iface.CodeThunderyShowers,
	}

	var today iface.Cond

	today.Time = time.Unix(int64(dp.Time), 0)

	today.Code = iface.CodeUnknown
	if val, ok := codemap[dp.Icon]; ok {
		today.Code = val
	}
	today.Desc = dp.Summary

	var todayTempC *float32
	todayTempC = dp.Temperature
	today.TempC = todayTempC

	if dp.ApparentTemperature != nil && *dp.ApparentTemperature >= 0 {
		//var todayApparentTemperature *float32
		//todayApparentTemperature = dp.ApparentTemperature
		today.FeelsLikeC = new(float32)
		*today.FeelsLikeC = *dp.ApparentTemperature
	}

	if dp.PrecipProbability >= 0 {
		var todayChanceOfRainPercent int
		todayChanceOfRainPercent = int(dp.PrecipProbability * float32(100))
		today.ChanceOfRainPercent = &todayChanceOfRainPercent
	}

	//(only defined on hourly and daily data points)
	if dp.PrecipAccumulation != nil && *dp.PrecipAccumulation >= 0 {
		today.PrecipM = new(float32)
		*today.PrecipM = *dp.PrecipAccumulation
	}

	if dp.Visibility != nil && *dp.Visibility >= 0 {
		today.VisibleDistM = new(float32)
		today.VisibleDistM = dp.Visibility
	}

	if dp.WindSpeed != nil && *dp.WindSpeed >= 0 {
		today.WindspeedKmph = new(float32)
		today.WindspeedKmph = dp.WindSpeed
	}

	//today.WindGustKmph = resp.Currently.WindSpeed

	if dp.WindBearing >= 0 {
		var todayWindBearing int
		todayWindBearing = int(dp.WindBearing * float32(360))
		today.WinddirDegree = &todayWindBearing
	}

	return today
}

func (c *forecastConfig) Setup() {
	flag.StringVar(&c.apiKey, "forecast-api-key", "", "forecast backend: the api `KEY` to use")
	flag.BoolVar(&c.debug, "forecast-debug", false, "forecast backend: print raw requests and responses")
}

func (c *forecastConfig) Fetch(location string, numdays int) iface.Data {
	var ret iface.Data

	if len(c.apiKey) == 0 {
		log.Fatal("No forecast.io API key specified.")
	}
	requri := fmt.Sprintf(forecastWuri, c.apiKey, location)

	if c.debug {
		log.Printf("Weather service: %s\n", requri)
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
		log.Println("Weather request:", requri)
		//log.Printf("Weather response: %s\n", body)
	}

	var resp forecastResponse
	if err = json.Unmarshal(body, &resp); err != nil {
		log.Println(err)
	}

	//log.Printf("Weather response: %v\n", resp)

	//log.Printf("Weather response: %v\n", resp.Currently)

	ret.Location = fmt.Sprintf("%f:%f", *resp.Latitude, *resp.Longitude)
	var reqLatLon iface.LatLon
	reqLatLon.Latitude = *resp.Latitude
	reqLatLon.Longitude = *resp.Longitude
	ret.GeoLoc = &reqLatLon

	ret.Current = resp.Currently.Convert(c)
	ret.Forecast = resp.Hourly.Convert(c)

	return ret
}

func init() {
	iface.AllBackends["forecast.io"] = &forecastConfig{}
}
