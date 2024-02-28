package backends

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/schachmat/wego/iface"
	"io"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"
)

type openWeatherConfig struct {
	apiKey string
	lang   string
	debug  bool
}

type openWeatherResponse struct {
	Cod  string `json:"cod"`
	City struct {
		Name    string `json:"name"`
		Country string `json:"country"`
		TimeZone int64 `json: "timezone"`
		// sunrise/sunset are once per call
		SunRise int64 `json: "sunrise"`
		SunSet int64 `json: "sunset"`
	} `json:"city"`
	List []dataBlock `json:"list"`
}

type dataBlock struct {
	Dt   int64 `json:"dt"`
	Main struct {
		TempC      float32 `json:"temp"`
		FeelsLikeC float32 `json:"feels_like"`
		Humidity   int     `json:"humidity"`
	} `json:"main"`

	Weather []struct {
		Description string `json:"description"`
		ID          int    `json:"id"`
	} `json:"weather"`

	Wind struct {
		Speed float32 `json:"speed"`
		Deg   float32 `json:"deg"`
	} `json:"wind"`

	Rain struct {
		MM3h float32 `json:"3h"`
	} `json:"rain"`
}

const (
	openweatherURI = "http://api.openweathermap.org/data/2.5/forecast?%s&appid=%s&units=metric&lang=%s"
)

func (c *openWeatherConfig) Setup() {
	flag.StringVar(&c.apiKey, "owm-api-key", "", "openweathermap backend: the api `KEY` to use")
	flag.StringVar(&c.lang, "owm-lang", "en", "openweathermap backend: the `LANGUAGE` to request from openweathermap")
	flag.BoolVar(&c.debug, "owm-debug", false, "openweathermap backend: print raw requests and responses")
}

func (c *openWeatherConfig) fetch(url string) (*openWeatherResponse, error) {
	res, err := http.Get(url)
	if c.debug {
		fmt.Printf("Fetching %s\n", url)
	}
	if err != nil {
		return nil, fmt.Errorf(" Unable to get (%s) %v", url, err)
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("Unable to read response body (%s): %v", url, err)
	}

	if c.debug {
		fmt.Printf("Response (%s):\n%s\n", url, string(body))
	}

	var resp openWeatherResponse
	if err = json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("Unable to unmarshal response (%s): %v\nThe json body is: %s", url, err, string(body))
	}
	if resp.Cod != "200" {
		return nil, fmt.Errorf("Erroneous response body: %s", string(body))
	}
	return &resp, nil
}

func (c *openWeatherConfig) parseDaily(dataInfo []dataBlock, numdays int) []iface.Day {
	var forecast []iface.Day
	var day *iface.Day

	for _, data := range dataInfo {
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

func (c *openWeatherConfig) parseCond(dataInfo dataBlock) (iface.Cond, error) {
	var ret iface.Cond
	codemap := map[int]iface.WeatherCode{
		200: iface.CodeThunderyShowers,
		201: iface.CodeThunderyShowers,
		210: iface.CodeThunderyShowers,
		230: iface.CodeThunderyShowers,
		231: iface.CodeThunderyShowers,
		202: iface.CodeThunderyHeavyRain,
		211: iface.CodeThunderyHeavyRain,
		212: iface.CodeThunderyHeavyRain,
		221: iface.CodeThunderyHeavyRain,
		232: iface.CodeThunderyHeavyRain,
		300: iface.CodeLightRain,
		301: iface.CodeLightRain,
		310: iface.CodeLightRain,
		311: iface.CodeLightRain,
		313: iface.CodeLightRain,
		321: iface.CodeLightRain,
		302: iface.CodeHeavyRain,
		312: iface.CodeHeavyRain,
		314: iface.CodeHeavyRain,
		500: iface.CodeLightShowers,
		501: iface.CodeLightShowers,
		502: iface.CodeHeavyShowers,
		503: iface.CodeHeavyShowers,
		504: iface.CodeHeavyShowers,
		511: iface.CodeLightSleet,
		520: iface.CodeLightShowers,
		521: iface.CodeLightShowers,
		522: iface.CodeHeavyShowers,
		531: iface.CodeHeavyShowers,
		600: iface.CodeLightSnow,
		601: iface.CodeLightSnow,
		602: iface.CodeHeavySnow,
		611: iface.CodeLightSleet,
		612: iface.CodeLightSleetShowers,
		615: iface.CodeLightSleet,
		616: iface.CodeLightSleet,
		620: iface.CodeLightSnowShowers,
		621: iface.CodeLightSnowShowers,
		622: iface.CodeHeavySnowShowers,
		701: iface.CodeFog,
		711: iface.CodeFog,
		721: iface.CodeFog,
		741: iface.CodeFog,
		731: iface.CodeUnknown, // sand, dust whirls
		751: iface.CodeUnknown, // sand
		761: iface.CodeUnknown, // dust
		762: iface.CodeUnknown, // volcanic ash
		771: iface.CodeUnknown, // squalls
		781: iface.CodeUnknown, // tornado
		800: iface.CodeSunny,
		801: iface.CodePartlyCloudy,
		802: iface.CodeCloudy,
		803: iface.CodeVeryCloudy,
		804: iface.CodeVeryCloudy,
		900: iface.CodeUnknown, // tornado
		901: iface.CodeUnknown, // tropical storm
		902: iface.CodeUnknown, // hurricane
		903: iface.CodeUnknown, // cold
		904: iface.CodeUnknown, // hot
		905: iface.CodeUnknown, // windy
		906: iface.CodeUnknown, // hail
		951: iface.CodeUnknown, // calm
		952: iface.CodeUnknown, // light breeze
		953: iface.CodeUnknown, // gentle breeze
		954: iface.CodeUnknown, // moderate breeze
		955: iface.CodeUnknown, // fresh breeze
		956: iface.CodeUnknown, // strong breeze
		957: iface.CodeUnknown, // high wind, near gale
		958: iface.CodeUnknown, // gale
		959: iface.CodeUnknown, // severe gale
		960: iface.CodeUnknown, // storm
		961: iface.CodeUnknown, // violent storm
		962: iface.CodeUnknown, // hurricane
	}

	ret.Code = iface.CodeUnknown
	ret.Desc = dataInfo.Weather[0].Description
	ret.Humidity = &(dataInfo.Main.Humidity)
	ret.TempC = &(dataInfo.Main.TempC)
	ret.FeelsLikeC = &(dataInfo.Main.FeelsLikeC)
	if &dataInfo.Wind.Deg != nil {
		p := int(dataInfo.Wind.Deg)
		ret.WinddirDegree = &p
	}
	if &(dataInfo.Wind.Speed) != nil && (dataInfo.Wind.Speed) > 0 {
		windSpeed := (dataInfo.Wind.Speed * 3.6)
		ret.WindspeedKmph = &(windSpeed)
	}
	if val, ok := codemap[dataInfo.Weather[0].ID]; ok {
		ret.Code = val
	}

	if &dataInfo.Rain.MM3h != nil {
		mmh := (dataInfo.Rain.MM3h / 1000) / 3
		ret.PrecipM = &mmh
	}

	ret.Time = time.Unix(dataInfo.Dt, 0)

	return ret, nil
}

func (c *openWeatherConfig) Fetch(location string, numdays int) iface.Data {
	var ret iface.Data
	loc := ""

	if len(c.apiKey) == 0 {
		log.Fatal("No openweathermap.org API key specified.\nYou have to register for one at https://home.openweathermap.org/users/sign_up")
	}
	if matched, err := regexp.MatchString(`^-?[0-9]*(\.[0-9]+)?,-?[0-9]*(\.[0-9]+)?$`, location); matched && err == nil {
		s := strings.Split(location, ",")
		loc = fmt.Sprintf("lat=%s&lon=%s", s[0], s[1])
	} else if matched, err = regexp.MatchString(`^[0-9].*`, location); matched && err == nil {
		loc = "zip=" + location
	} else {
		loc = "q=" + location
	}

	resp, err := c.fetch(fmt.Sprintf(openweatherURI, loc, c.apiKey, c.lang))
	if err != nil {
		log.Fatalf("Failed to fetch weather data: %v\n", err)
	}
	ret.Current, err = c.parseCond(resp.List[0])
	ret.Location = fmt.Sprintf("%s, %s", resp.City.Name, resp.City.Country)

	if err != nil {
		log.Fatalf("Failed to fetch weather data: %v\n", err)
	}
	ret.Forecast = c.parseDaily(resp.List, numdays)

	// add in the sunrise/sunset information to the first day
	// these maybe should deal with resp.City.TimeZone
	if len(ret.Forecast) > 0 {
		ret.Forecast[0].Astronomy.Sunrise = time.Unix(resp.City.SunRise, 0)
		ret.Forecast[0].Astronomy.Sunset = time.Unix(resp.City.SunSet, 0)
	}

	return ret
}

func init() {
	iface.AllBackends["openweathermap"] = &openWeatherConfig{}
}
