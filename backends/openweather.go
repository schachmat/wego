package backends

import (
	"flag"
	//"fmt"
	"encoding/json"
	"fmt"
	"github.com/schachmat/wego/iface"
	"io/ioutil"
	"net/http"
	"strings"
	//"time"
	"log"
	//"os"
	"time"
)

type openWeatherConfig struct {
	apiKey string
	lang   string
	debug  bool
}

type openWeatherResponse struct {
	Cod     string            `json:"cod"`
	Message float64           `json:"message"`
	Cnt     int               `json:"cnt"`
	City    cityResponseBlock `json:"city"`
	List    []listBlock       `json:"list"`
}

type listBlock struct {
	Dt     int64  `json:"dt"`
	Dt_txt string `json:"dt_txt"`
	//Main  section
	Main struct {
		Temp       float32 `json:"temp"`
		Temp_min   float32 `json:"temp_min"`
		Temp_max   float32 `json:"temp_max"`
		Pressure   float32 `json:"pressure"`
		Sea_level  float32 `json:"sea_level"`
		Grnd_level float32 `json:"grnd_level"`
		Humidity   int     `json:"humidity"`
		Temp_kf    float32 `json:"temp_kf"`
	} `json:"main"`

	Clouds struct {
		All int `json:"all"`
	} `json:"clouds"`

	Weather []struct {
		Description string `json:"description"`
		Icon        string `json:"icon"`
		ID          int    `json:"id"`
		Main        string `json:"main"`
	} `json:"weather"`

	Wind struct {
		Speed float32 `json:"speed"`
		Deg   float32 `json:"deg"`
	} `json:"wind"`

	Rain struct {
		MM3h float32 `json:"3h"`
	} `json:"rain"`
}

type cityResponseBlock struct {
	Id      int    `json:"id"`
	Name    string `json:"name"`
	Country string `json:"country"`
}

const (
	openweatherUri = "http://api.openweathermap.org/data/2.5/forecast?lat=%s&lon=%s&appid=%s&units=metric&lang=%s"
)

func (ow *openWeatherConfig) Setup() {
	flag.StringVar(&ow.apiKey, "openweather-api-key", "", "openweather backend: the api `KEY` to use")
	flag.StringVar(&ow.lang, "openweather-lang", "en", "openweather backend: the `LANGUAGE` to request to openweather")
	flag.BoolVar(&ow.debug, "openweather-debug", false, "openweather backend: print raw requests and responses")
}

func (ow *openWeatherConfig) fetch(url string) (*openWeatherResponse, error) {
	res, err := http.Get(url)
	if ow.debug {
		fmt.Printf("Fetching for %s \n", url)
	}
	if err != nil {
		return nil, fmt.Errorf(" Unable to get (%s) %v", url, err)
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("Unable to read response body (%s): %v", url, err)
	}

	if ow.debug {
		fmt.Printf("Response (%s) %s", url, string(body))
	}

	var resp openWeatherResponse
	if err = json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("Unable to unmarshal response (%s): %v\nThe json body is: %s", url, err, string(body))
	}
	return &resp, nil
}

func (ow *openWeatherConfig) parseDaily(dataInfo []listBlock, numdays int) []iface.Day {
	var openWeather []iface.Day
	var day *iface.Day

	for _, data := range dataInfo {
		slot, err := ow.parseCond(data)
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
			openWeather = append(openWeather, *day)
			if len(openWeather) >= numdays {
				break
			}
			day = new(iface.Day)
			day.Date = slot.Time
			day.Slots = append(day.Slots, slot)
		}

	}
	return openWeather
}

func (ow *openWeatherConfig) parseCond(dataInfo listBlock) (iface.Cond, error) {
	var ret iface.Cond
	codemap := map[int]iface.WeatherCode{
		500: iface.CodeLightRain,
		501: iface.CodeHeavyRain,
		502: iface.CodeHeavyRain,
		503: iface.CodeHeavyRain,
		800: iface.CodeSunny,
		802: iface.CodePartlyCloudy,
		803: iface.CodePartlyCloudy,
		804: iface.CodePartlyCloudy,
	}
	ret.Code = iface.CodeUnknown
	ret.Desc = dataInfo.Weather[0].Main
	ret.Humidity = &(dataInfo.Main.Humidity)
	ret.TempC = &(dataInfo.Main.Temp)
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
		mmh := dataInfo.Rain.MM3h / 3
		ret.PrecipM = &mmh
	}

	ret.Time = time.Unix(dataInfo.Dt, 0)

	return ret, nil
}

func (ow *openWeatherConfig) Fetch(location string, numdays int) iface.Data {
	var ret iface.Data
	s := strings.Split(location, ",")

	lat, lon := s[0], s[1]

	resp, err := ow.fetch(fmt.Sprintf(openweatherUri, lat, lon, ow.apiKey, ow.lang))
	if err != nil {
		log.Println("Cant fetch today")
	}
	ret.Current, err = ow.parseCond(resp.List[0])
	ret.Location = fmt.Sprintf("%s", resp.City.Name)

	if err != nil {
		log.Fatalf("Failed to fetch weather data: %v\n", err)
	}
	ret.Forecast = ow.parseDaily(resp.List, numdays)
	return ret
}

func init() {
	iface.AllBackends["openweather"] = &openWeatherConfig{}
}
