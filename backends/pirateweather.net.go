package backends

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/schachmat/wego/iface"
)

const (
	pirateweatherURI = "https://api.pirateweather.net/forecast/"
)

type PirateweatherConfig struct {
	apiKey string
	debug  bool
}

func (c *PirateweatherConfig) Setup() {
	flag.StringVar(&c.apiKey, "pirateweather-api-key", "", "pirateweather backend: the api `KEY` to use")
	flag.BoolVar(&c.debug, "pirateweather-debug", true, "pirateweather backend: print raw request and responses")
}

func (c *PirateweatherConfig) Fetch(location string, numdays int) iface.Data {
	if c.debug {
		log.Printf("pirateweather location %v", location)
	}

	res := iface.Data{}
	reqURI := fmt.Sprintf("%s/%s/%s", pirateweatherURI, c.apiKey, location)
	apiRes, err := http.Get(reqURI)
	if err != nil || apiRes.StatusCode != 200 {
		panic(err)
	}
	defer apiRes.Body.Close()

	body, err := io.ReadAll(apiRes.Body)
	if err != nil {
		panic(err)
	}

	if c.debug {
		log.Println("pirateweather request:", reqURI)
		data, _ := json.MarshalIndent(body, "", "\t")
		log.Println("pirateweather response:", string(data))
	}

	weatherData := &PirateweatherWeather{}
	if err := json.Unmarshal(body, weatherData); err != nil {
		panic(err)
	}

	// TODO: parse weatherData in iface.Data

	return res
}

func init() {
	iface.AllBackends["pirateweather.net"] = &PirateweatherConfig{}
}

type PirateweatherWeather struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Timezone  string  `json:"timezone"`
	Offset    string  `json:"offset"`
	Currently struct {
		Time                uint    `json:"time"`
		Summary             string  `json:"summary"`
		Icon                string  `json:"icon"`
		PrecipIntensity     float64 `json:"precipIntensity"`
		PrecipType          string  `json:"precipType"`
		Temperature         float64 `json:"temperature"`
		ApparentTemperature float64 `json:"apparentTemperature"`
		DewPoint            float64 `json:"dewPoint"`
		Pressure            float64 `json:"pressure"`
		WindSpeed           float64 `json:"windSpeed"`
		WindGust            float64 `json:"windGust"`
		WindBearing         uint    `json:"windBearing"`
		CloudCover          float64 `json:"cloudCover"`
	} `json:"currently"`
	Hourly struct {
		Data []struct {
			Time                uint    `json:"time"`
			Icon                string  `json:"icon"`
			Summary             string  `json:"summary"`
			PrecipAccumulation  float64 `json:"precipAccumulation"`
			PrecipType          string  `json:"precipType"`
			Temperature         float64 `json:"temperature"`
			ApparentTemperature string  `json:"apparentTemperature"`
			DewPoint            float64 `json:"dewPoint"`
			Pressure            float64 `json:"pressure"`
			WindSpeed           float64 `json:"windSpeed"`
			WindGust            float64 `json:"windGust"`
			WindBearing         uint    `json:"windBearing"`
			CloudCover          float64 `json:"cloudCover"`
			SnowAccumulation    float64 `json:"snowAccumulation"`
		} `json:"data"`
	} `json:"hourly"`
	Daily struct {
		Data []struct {
			Time                        uint    `json:"time"`
			Icon                        string  `json:"icon"`
			Summary                     string  `json:"summary"`
			SunriseTime                 uint    `json:"sunriseTime"`
			SunsetTime                  uint    `json:"sunsetTime"`
			MoonPhase                   float64 `json:"moonPhase"`
			PrecipAccumulation          float64 `json:"precipAccumulation"`
			PrecipType                  string  `json:"precipType"`
			TemperatureHigh             float64 `json:"temperatureHigh"`
			TemperatureHighTime         uint    `json:"temperatureHighTime"`
			TemperatureLow              float64 `json:"temperatureLow"`
			TemperatureLowTime          uint    `json:"temperatureLowTime"`
			ApparentTemperatureHigh     float64 `json:"apparentTemperatureHigh"`
			ApparentTemperatureHighTime uint    `json:"apparentTemperatureHighTime"`
			ApparentTemperatureLow      float64 `json:"apparentTemperatureLow"`
			ApparentTemperatureLowTime  float64 `json:"apparentTemperatureLowTime"`
			Pressure                    float64 `json:"pressure"`
			WindSpeed                   float64 `json:"windSpeed"`
			WindGust                    float64 `json:"windGust"`
			WindGustTime                uint    `json:"windGustTime"`
			WindBearing                 uint    `json:"windBearing"`
			CloudCover                  float64 `json:"cloudCover"`
			TemperatureMin              float64 `json:"temperatureMin"`
			TemperatureMinTime          uint    `json:"temperatureMinTime"`
			TemperatureMax              float64 `json:"temperatureMax"`
			TemperatureMaxTime          uint    `json:"temperatureMaxTime"`
			ApparentTemperatureMin      float64 `json:"apparentTemperatureMin"`
			ApparentTemperatureMinTime  float64 `json:"apparentTemperatureMinTime"`
			ApparentTemperatureMax      float64 `json:"apparentTemperatureMax"`
			ApparentTemperatureMaxTime  uint    `json:"apparentTemperatureMaxTime"`
			SnowAccumulation            float64 `json:"snowAccumulation"`
		} `json:"data"`
	} `json:"daily"`
	Flags struct {
		Sources        string `json:"sources"`
		NearestStation uint   `json:"nearest-station"`
		Units          string `json:"units"`
		Version        string `json:"version"`
		SourceIDX      map[string]struct {
			X    uint    `json:"x"`
			Y    uint    `json:"y"`
			Lat  float64 `json:"lat"`
			Long float64 `json:"long"`
		} `json:"sourceIDX"`
		ProcessTime uint `json:"processTime"`
	} `json:"flags"`
}
