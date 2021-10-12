package backends

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/ringsaturn/wego/iface"
)

const (
	CAIYUNAPI       = "http://api.caiyunapp.com/v2.6/%s/%s/weather?lang=%s&alert=true&unit=metric:v2"
	CAIYUNDATE_TMPL = "2006-01-02T15:04-07:00"
)

type CaiyunConfig struct {
	apiKey string
	lang   string
	// debug  bool
	tz *time.Location
}

func (c *CaiyunConfig) Setup() {
	flag.StringVar(&c.apiKey, "caiyun-api-key", "", "forecast backend: the api `KEY` to use")
	flag.StringVar(&c.lang, "caiyun-lang", "en", "forecast backend: the `LANGUAGE` to request from caiyunapp.com/")
	// flag.BoolVar(&c.debug, "forecast-debug", false, "forecast backend: print raw requests and responses")
}

var SkyconToIfaceCode map[string]iface.WeatherCode

func init() {
	//SkyconToIfaceCode["CLEAR_DAY"] = iface.CodeSunny
	SkyconToIfaceCode = map[string]iface.WeatherCode{
		"CLEAR_DAY":           iface.CodeSunny,
		"CLEAR_NIGHT":         iface.CodeSunny,
		"PARTLY_CLOUDY_DAY":   iface.CodePartlyCloudy,
		"PARTLY_CLOUDY_NIGHT": iface.CodePartlyCloudy,
		"CLOUDY":              iface.CodeCloudy,
		"LIGHT_HAZE":          iface.CodeUnknown,
		"MODERATE_HAZE":       iface.CodeUnknown,
		"HEAVY_HAZE":          iface.CodeUnknown,
		"LIGHT_RAIN":          iface.CodeLightRain,
		"MODERATE_RAIN":       iface.CodeLightRain,
		"HEAVY_RAIN":          iface.CodeHeavyRain,
		"STORM_RAIN":          iface.CodeHeavyRain,
		"FOG":                 iface.CodeFog,
		"LIGHT_SNOW":          iface.CodeLightSnow,
		"MODERATE_SNOW":       iface.CodeLightSnow,
		"HEAVY_SNOW":          iface.CodeHeavySnow,
		"STORM_SNOW":          iface.CodeHeavySnow,
		"DUST":                iface.CodeUnknown,
		"SAND":                iface.CodeUnknown,
		"WIND":                iface.CodeUnknown,
	}
}

func (c *CaiyunConfig) Fetch(location string, numdays int) iface.Data {
	res := iface.Data{}
	url := fmt.Sprintf(CAIYUNAPI, c.apiKey, location, c.lang)
	resp, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	weatherData := &CaiyunWeather{}
	if err := json.Unmarshal(body, weatherData); err != nil {
		panic(err)
	}
	res.Current.Desc = weatherData.Result.ForecastKeypoint
	res.Current.TempC = func() *float32 {
		x := float32(weatherData.Result.Realtime.Temperature)
		return &x
	}()
	if code, ok := SkyconToIfaceCode[weatherData.Result.Realtime.Skycon]; ok {
		res.Current.Code = code
	} else {
		res.Current.Code = iface.CodeUnknown
	}
	if len(weatherData.Result.Alert.Adcodes) != 0 {
		adcodes := weatherData.Result.Alert.Adcodes
		if len(adcodes) == 3 {
			res.Location = adcodes[1].Name + adcodes[2].Name
		}
		if len(adcodes) == 2 {
			res.Location = adcodes[0].Name + adcodes[1].Name
		}
	} else {
		res.Location = "第三红岸基地"
	}
	res.Current.WinddirDegree = func() *int {
		x := int(weatherData.Result.Realtime.Wind.Direction)
		return &x
	}()
	res.Current.WindspeedKmph = func() *float32 {
		x := float32(weatherData.Result.Realtime.Wind.Speed)
		return &x
	}()
	res.Current.PrecipM = func() *float32 {
		x := float32(weatherData.Result.Realtime.Precipitation.Local.Intensity)
		return &x
	}()
	dailyDataSlice := []iface.Day{}
	for i := 0; i < numdays; i++ {
		weatherDailyData := weatherData.Result.Daily

		dailyData := iface.Day{
			Date: func() time.Time {
				x, err := time.Parse(CAIYUNDATE_TMPL, weatherDailyData.Temperature[i].Date)
				if err != nil {
					panic(err)
				}
				return x
			}(),
			// Astronomy: iface.Astro{
			// 	Sunrise: func() time.Time {
			// 		fmt.Println("weatherDailyData.Astro[i].Sunrise.Time", weatherDailyData.Astro[i].Sunrise.Time)
			// 		x, err := time.Parse(CAIYUNDATE_TMPL, weatherDailyData.Astro[i].Sunrise.Time)
			// 		if err != nil {
			// 			panic(err)
			// 		}
			// 		return x
			// 	}(),
			// 	Sunset: func() time.Time {
			// 		x, err := time.Parse(CAIYUNDATE_TMPL, weatherDailyData.Astro[i].Sunset.Time)
			// 		if err != nil {
			// 			panic(err)
			// 		}
			// 		return x
			// 	}(),
			// },
			Slots: []iface.Cond{},
		}

		// Morning
		dailyData.Slots = append(dailyData.Slots, iface.Cond{
			TempC: func() *float32 {
				x := float32(weatherDailyData.Temperature08H20H[i].Avg)
				return &x
			}(),
			Time: func() time.Time {
				x, err := time.Parse(CAIYUNDATE_TMPL, weatherDailyData.Temperature[i].Date)
				if err != nil {
					panic(err)
				}
				return x
			}(),
			Code: func() iface.WeatherCode {
				if code, ok := SkyconToIfaceCode[weatherDailyData.Skycon08H20H[i].Value]; ok {
					return code
				} else {
					return iface.CodeUnknown
				}
			}(),
			PrecipM: func() *float32 {
				x := float32(weatherDailyData.Precipitation[i].Avg) / 1000
				return &x
			}(),
		})
		// Noon
		dailyData.Slots = append(dailyData.Slots, iface.Cond{
			TempC: func() *float32 {
				x := float32(weatherDailyData.Temperature08H20H[i].Avg)
				return &x
			}(),
			Time: func() time.Time {
				x, err := time.Parse(CAIYUNDATE_TMPL, weatherDailyData.Temperature[i].Date)
				if err != nil {
					panic(err)
				}
				return x
			}(),
			Code: func() iface.WeatherCode {
				if code, ok := SkyconToIfaceCode[weatherDailyData.Skycon08H20H[i].Value]; ok {
					return code
				} else {
					return iface.CodeUnknown
				}
			}(),
			PrecipM: func() *float32 {
				x := float32(weatherDailyData.Precipitation[i].Avg) / 1000
				return &x
			}(),
		})
		// Evening
		dailyData.Slots = append(dailyData.Slots, iface.Cond{
			TempC: func() *float32 {
				x := float32(weatherDailyData.Temperature08H20H[i].Avg)
				return &x
			}(),
			Time: func() time.Time {
				x, err := time.Parse(CAIYUNDATE_TMPL, weatherDailyData.Temperature[i].Date)
				if err != nil {
					panic(err)
				}
				return x
			}(),
			Code: func() iface.WeatherCode {
				if code, ok := SkyconToIfaceCode[weatherDailyData.Skycon08H20H[i].Value]; ok {
					return code
				} else {
					return iface.CodeUnknown
				}
			}(),
			PrecipM: func() *float32 {
				x := float32(weatherDailyData.Precipitation[i].Avg) / 1000
				return &x
			}(),
		})
		// Night
		dailyData.Slots = append(dailyData.Slots, iface.Cond{
			TempC: func() *float32 {
				x := float32(weatherDailyData.Temperature20H32H[i].Avg)
				return &x
			}(),
			Time: func() time.Time {
				x, err := time.Parse(CAIYUNDATE_TMPL, weatherDailyData.Temperature[i].Date)
				if err != nil {
					panic(err)
				}
				return x
			}(),
			Code: func() iface.WeatherCode {
				if code, ok := SkyconToIfaceCode[weatherDailyData.Skycon20H32H[i].Value]; ok {
					return code
				} else {
					return iface.CodeUnknown
				}
			}(),
			PrecipM: func() *float32 {
				x := float32(weatherDailyData.Precipitation[i].Avg) / 1000
				return &x
			}(),
		})

		dailyDataSlice = append(dailyDataSlice, dailyData)
	}
	res.Forecast = dailyDataSlice

	res.GeoLoc = &iface.LatLon{
		Latitude:  float32(weatherData.Location[0]),
		Longitude: float32(weatherData.Location[1]),
	}
	return res
}

func init() {
	iface.AllBackends["caiyunapp.com"] = &CaiyunConfig{}
}
