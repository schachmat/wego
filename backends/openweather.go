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
	Dt     int    `json:"dt"`
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
		Speed float64 `json:"speed"`
		Deg   float64 `json:"deg"`
	} `json:"wind"`

	Rain struct {
		Chance float32 `json:"3h"`
	} `json:"rain"`
}

type cityResponseBlock struct {
	Id      int    `json:"id"`
	Name    string `json:"name"`
	Country string `json:"country"`
}

const (
	// http://api.openweathermap.org/data/2.5/forecast?lat=35&lon=139&appid=ad837f4f16d346d3cf15ad700749fd3e
	//openweatherUri = "http://api.openweathermap.org/data/2.5/forecast?lat=%s&lon=%s&appid=%s&appid=ad837f4f16d346d3cf15ad700749fd3e"
	openweatherUri = "http://api.openweathermap.org/data/2.5/forecast?lat=%s&lon=%s&appid=%s"
)

func (ow *openWeatherConfig) Setup() {
	flag.StringVar(&ow.apiKey, "openweather-api-key", "", "openweather backend: the api `KEY` to use")
	flag.StringVar(&ow.lang, "openweather-lang", "en", "openweather backend: the `LANGUAGE` to request from forecast.io")
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

func (ow *openWeatherConfig) parse(dataInfo []listBlock, numdays int) []iface.Day {
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
		day.Slots = append(day.Slots, slot)
	}
	return append(openWeather, *day)
}

func (ow *openWeatherConfig) parseCond(dataInfo listBlock) (iface.Cond, error) {
	var ret iface.Cond

	ret.Desc = dataInfo.Weather[0].Description
	ret.Humidity = &dataInfo.Main.Humidity
	//*ret.TempC = dataInfo.Main.Temp
	//*ret.ChanceOfRainPercent = 2

	return ret, nil
}

func (ow *openWeatherConfig) fetchToday(location string) ([]iface.Cond, error) {
	s := strings.Split(location, ",")
	var ret []iface.Cond
	lat, lon := s[0], s[1]
	urlToFetch := fmt.Sprintf(openweatherUri, lat, lon, ow.apiKey)
	resp, err := ow.fetch(urlToFetch)
	if err != nil {
		return nil, fmt.Errorf("Failed to fetch %s", urlToFetch)
	}

	parsedDay := ow.parse(resp.List, 1)
	fmt.Println("%v", parsedDay)
	return ret, nil
}

func (ow *openWeatherConfig) Fetch(location string, numdays int) iface.Data {
	var ret iface.Data
	//todayChan := make(chan []iface.Cond)

	_, err := ow.fetchToday(location)
	if err != nil {
		log.Fatal(err)
	}
	return ret
}

func init() {
	iface.AllBackends["openweather"] = &openWeatherConfig{}
}
