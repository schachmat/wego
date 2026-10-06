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

type yrNoConfig struct {
	userAgent string
	debug     bool
}

type yrNoResponse struct {
	Properties struct {
		Timeseries []timeSeriesEntry `json:"timeseries"`
	} `json:"properties"`
}

type timeSeriesEntry struct {
	Dt   string `json:"time"`
	Data struct {
		Instant struct {
			Details struct {
				AirPressure       float32 `json:"air_pressure_at_sea_level"`
				TempC             float32 `json:"air_temperature"`
				RelativeHumidity  float32 `json:"relative_humidity"`
				WindSpeed         float32 `json:"wind_speed"`
				WindFromDirection float32 `json:"wind_from_direction"`
			} `json:"details"`
		} `json:"instant"`
		NextOneHour struct {
			Summary struct {
				SymbolCode string `json:"symbol_code"`
			} `json:"summary"`
			Details struct {
				Precipitation float32 `json:"precipitation_amount"`
			} `json:"details"`
		} `json:"next_1_hours"`
	} `json:"data"`
}

const (
	yrNoURI = "https://api.met.no/weatherapi/locationforecast/2.0/compact?%s"
)

func (c *yrNoConfig) Setup() {
	flag.StringVar(&c.userAgent, "yrno-user-agent", "", "yr.no backend: the user agent to use. See https://docs.api.met.no/doc/TermsOfService.html for details")
	flag.BoolVar(&c.debug, "yrno-debug", false, "yr.no backend: print raw requests and responses")
}

func (c *yrNoConfig) fetch(url string) (*yrNoResponse, error) {
	client := &http.Client{}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatalln(err)
	}

	req.Header.Set("User-Agent", c.userAgent)

	res, err := client.Do(req)
	if c.debug {
		fmt.Printf("Fetching %s\n", url)
	}
	if err != nil {
		return nil, fmt.Errorf(" Unable to get (%s) %v", url, err)
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("unable to read response body (%s): %v", url, err)
	}

	if c.debug {
		fmt.Printf("Response (%s):\n%s\n", url, string(body))
	}

	var resp yrNoResponse
	if err = json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("unable to unmarshal response (%s): %v\nThe json body is: %s", url, err, string(body))
	}
	return &resp, nil
}

func (c *yrNoConfig) parseDaily(entries []timeSeriesEntry, numdays int) []iface.Day {
	var forecast []iface.Day
	var day *iface.Day

	for _, data := range entries {
		slot, err := c.parseCond(data)
		if err != nil {
			log.Println("Error parsing hourly weather condition:", err)
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
	return forecast
}

func (c *yrNoConfig) parseCond(entry timeSeriesEntry) (iface.Cond, error) {
	var ret iface.Cond
	// descriptions from https://github.com/metno/weathericons/blob/main/weather/legend.csv
	descriptionMap := map[string]string{
		"clearsky":                     "Clear sky",
		"fair":                         "Fair",
		"partlycloudy":                 "Partly cloudy",
		"cloudy":                       "Cloudy",
		"lightrainshowers":             "Light rain showers",
		"rainshowers":                  "Rain showers",
		"heavyrainshowers":             "Heavy rain showers",
		"lightrainshowersandthunder":   "Light rain showers and thunder",
		"rainshowersandthunder":        "Rain showers and thunder",
		"heavyrainshowersandthunder":   "Heavy rain showers and thunder",
		"lightsleetshowers":            "Light sleet showers",
		"sleetshowers":                 "Sleet showers",
		"heavysleetshowers":            "Heavy sleet showers",
		"lightssleetshowersandthunder": "Light sleet showers and thunder",
		"sleetshowersandthunder":       "Sleet showers and thunder",
		"heavysleetshowersandthunder":  "Heavy sleet showers and thunder",
		"lightsnowshowers":             "Light snow showers",
		"snowshowers":                  "Snow showers",
		"heavysnowshowers":             "Heavy snow showers",
		"lightssnowshowersandthunder":  "Light snow showers and thunder",
		"snowshowersandthunder":        "Snow showers and thunder",
		"heavysnowshowersandthunder":   "Heavy snow showers and thunder",
		"lightrain":                    "Light rain",
		"rain":                         "Rain",
		"heavyrain":                    "Heavy rain",
		"lightrainandthunder":          "Light rain and thunder",
		"rainandthunder":               "Rain and thunder",
		"heavyrainandthunder":          "Heavy rain and thunder",
		"lightsleet":                   "Light sleet",
		"sleet":                        "Sleet",
		"heavysleet":                   "Heavy sleet",
		"lightsleetandthunder":         "Light sleet and thunder",
		"sleetandthunder":              "Sleet and thunder",
		"heavysleetandthunder":         "Heavy sleet and thunder",
		"lightsnow":                    "Light snow",
		"snow":                         "Snow",
		"heavysnow":                    "Heavy snow",
		"lightsnowandthunder":          "Light snow and thunder",
		"snowandthunder":               "Snow and thunder",
		"heavysnowandthunder":          "Heavy snow and thunder",
		"fog":                          "Fog",
	}

	// codes from https://api.met.no/weatherapi/locationforecast/2.0/swagger
	codemap := map[string]iface.WeatherCode{
		"clearsky_day":                               iface.CodeSunny,
		"clearsky_night":                             iface.CodeSunny,
		"clearsky_polartwilight":                     iface.CodeSunny,
		"fair_day":                                   iface.CodeSunny,
		"fair_night":                                 iface.CodeCloudy,
		"fair_polartwilight":                         iface.CodeCloudy,
		"lightssnowshowersandthunder_day":            iface.CodeThunderySnowShowers,
		"lightssnowshowersandthunder_night":          iface.CodeThunderySnowShowers,
		"lightssnowshowersandthunder_polartwilight":  iface.CodeThunderySnowShowers,
		"lightsnowshowers_day":                       iface.CodeLightSnowShowers,
		"lightsnowshowers_night":                     iface.CodeLightSnowShowers,
		"lightsnowshowers_polartwilight":             iface.CodeLightSnowShowers,
		"heavyrainandthunder":                        iface.CodeThunderyHeavyRain,
		"heavysnowandthunder":                        iface.CodeThunderySnowShowers,
		"rainandthunder":                             iface.CodeThunderyHeavyRain,
		"heavysleetshowersandthunder_day":            iface.CodeThunderySnowShowers,
		"heavysleetshowersandthunder_night":          iface.CodeThunderySnowShowers,
		"heavysleetshowersandthunder_polartwilight":  iface.CodeThunderySnowShowers,
		"heavysnow":                                  iface.CodeHeavySnow,
		"heavyrainshowers_day":                       iface.CodeHeavyRain,
		"heavyrainshowers_night":                     iface.CodeHeavyRain,
		"heavyrainshowers_polartwilight":             iface.CodeHeavyRain,
		"lightsleet":                                 iface.CodeLightSleet,
		"heavyrain":                                  iface.CodeHeavyRain,
		"lightrainshowers_day":                       iface.CodeLightShowers,
		"lightrainshowers_night":                     iface.CodeLightShowers,
		"lightrainshowers_polartwilight":             iface.CodeLightShowers,
		"heavysleetshowers_day":                      iface.CodeHeavySnowShowers,
		"heavysleetshowers_night":                    iface.CodeHeavySnowShowers,
		"heavysleetshowers_polartwilight":            iface.CodeHeavySnowShowers,
		"lightsleetshowers_day":                      iface.CodeLightSleetShowers,
		"lightsleetshowers_night":                    iface.CodeLightSleetShowers,
		"lightsleetshowers_polartwilight":            iface.CodeLightSleetShowers,
		"snow":                                       iface.CodeLightSnow,
		"heavyrainshowersandthunder_day":             iface.CodeThunderyHeavyRain,
		"heavyrainshowersandthunder_night":           iface.CodeThunderyHeavyRain,
		"heavyrainshowersandthunder_polartwilight":   iface.CodeThunderyHeavyRain,
		"snowshowers_day":                            iface.CodeHeavySnowShowers,
		"snowshowers_night":                          iface.CodeHeavySnowShowers,
		"snowshowers_polartwilight":                  iface.CodeHeavySnowShowers,
		"fog":                                        iface.CodeFog,
		"snowshowersandthunder_day":                  iface.CodeThunderySnowShowers,
		"snowshowersandthunder_night":                iface.CodeThunderySnowShowers,
		"snowshowersandthunder_polartwilight":        iface.CodeThunderySnowShowers,
		"lightsnowandthunder":                        iface.CodeThunderySnowShowers,
		"heavysleetandthunder":                       iface.CodeThunderySnowShowers,
		"lightrain":                                  iface.CodeLightRain,
		"rainshowersandthunder_day":                  iface.CodeThunderyShowers,
		"rainshowersandthunder_night":                iface.CodeThunderyShowers,
		"rainshowersandthunder_polartwilight":        iface.CodeThunderyShowers,
		"rain":                                       iface.CodeHeavyRain,
		"lightsnow":                                  iface.CodeLightSnow,
		"lightrainshowersandthunder_day":             iface.CodeThunderyShowers,
		"lightrainshowersandthunder_night":           iface.CodeThunderyShowers,
		"lightrainshowersandthunder_polartwilight":   iface.CodeThunderyShowers,
		"heavysleet":                                 iface.CodeHeavySnowShowers,
		"sleetandthunder":                            iface.CodeThunderySnowShowers,
		"lightrainandthunder":                        iface.CodeThunderyHeavyRain,
		"sleet":                                      iface.CodeLightSleet,
		"lightssleetshowersandthunder_day":           iface.CodeThunderySnowShowers,
		"lightssleetshowersandthunder_night":         iface.CodeThunderySnowShowers,
		"lightssleetshowersandthunder_polartwilight": iface.CodeThunderySnowShowers,
		"lightsleetandthunder":                       iface.CodeThunderySnowShowers,
		"partlycloudy_day":                           iface.CodePartlyCloudy,
		"partlycloudy_night":                         iface.CodePartlyCloudy,
		"partlycloudy_polartwilight":                 iface.CodePartlyCloudy,
		"sleetshowersandthunder_day":                 iface.CodeThunderySnowShowers,
		"sleetshowersandthunder_night":               iface.CodeThunderySnowShowers,
		"sleetshowersandthunder_polartwilight":       iface.CodeThunderySnowShowers,
		"rainshowers_day":                            iface.CodeHeavyShowers,
		"rainshowers_night":                          iface.CodeHeavyShowers,
		"rainshowers_polartwilight":                  iface.CodeHeavyShowers,
		"snowandthunder":                             iface.CodeThunderySnowShowers,
		"sleetshowers_day":                           iface.CodeLightSleetShowers,
		"sleetshowers_night":                         iface.CodeLightSleetShowers,
		"sleetshowers_polartwilight":                 iface.CodeLightSleetShowers,
		"cloudy":                                     iface.CodeCloudy,
		"heavysnowshowersandthunder_day":             iface.CodeThunderySnowShowers,
		"heavysnowshowersandthunder_night":           iface.CodeThunderySnowShowers,
		"heavysnowshowersandthunder_polartwilight":   iface.CodeThunderySnowShowers,
		"heavysnowshowers_day":                       iface.CodeHeavySnowShowers,
		"heavysnowshowers_night":                     iface.CodeHeavySnowShowers,
		"heavysnowshowers_polartwilight":             iface.CodeHeavySnowShowers,
	}

	ret.Code = iface.CodeUnknown
	ret.Desc = entry.Data.NextOneHour.Summary.SymbolCode
	relHum := int(entry.Data.Instant.Details.RelativeHumidity)
	ret.Humidity = &(relHum)
	ret.TempC = &entry.Data.Instant.Details.TempC
	dir := int(entry.Data.Instant.Details.WindFromDirection)
	ret.WinddirDegree = &(dir)
	windSpeed := entry.Data.Instant.Details.WindSpeed * 3.6
	ret.WindspeedKmph = &(windSpeed)

	if val, ok := codemap[entry.Data.NextOneHour.Summary.SymbolCode]; ok {
		ret.Code = val
	}
	codeWithoutSuffix := entry.Data.NextOneHour.Summary.SymbolCode
	codeWithoutSuffix = strings.ReplaceAll(codeWithoutSuffix, "_day", "")
	codeWithoutSuffix = strings.ReplaceAll(codeWithoutSuffix, "_night", "")
	codeWithoutSuffix = strings.ReplaceAll(codeWithoutSuffix, "_polartwilight", "")
	if val, ok := descriptionMap[codeWithoutSuffix]; ok {
		ret.Desc = val
	}
	precipM := entry.Data.NextOneHour.Details.Precipitation / 1000.
	ret.PrecipM = &precipM
	ret.Time, _ = time.Parse(time.RFC3339, entry.Dt)
	return ret, nil
}

func (c *yrNoConfig) Fetch(location string, numdays int) iface.Data {
	var ret iface.Data
	loc := ""

	if len(c.userAgent) == 0 {
		log.Fatal("yr.no: No user agent specified.\n")
	}
	if matched, err := regexp.MatchString(`^-?[0-9]*(\.[0-9]+)?,-?[0-9]*(\.[0-9]+)?$`, location); matched && err == nil {
		s := strings.Split(location, ",")
		loc = fmt.Sprintf("lat=%s&lon=%s", s[0], s[1])
		ret.Location = location
	}

	resp, err := c.fetch(fmt.Sprintf(yrNoURI, loc))
	if err != nil {
		log.Fatalf("Failed to fetch weather data: %v\n", err)
	}
	ret.Current, err = c.parseCond(resp.Properties.Timeseries[0])

	if err != nil {
		log.Fatalf("Failed to fetch weather data: %v\n", err)
	}

	if numdays == 0 {
		return ret
	}
	ret.Forecast = c.parseDaily(resp.Properties.Timeseries, numdays)
	return ret
}

func init() {
	iface.AllBackends["yr.no"] = &yrNoConfig{}
}
