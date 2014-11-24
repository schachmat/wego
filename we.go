package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	_ "net/http"
	_ "net/url"
	"os"
	"os/user"
	"path"
	"strings"
)

type Config struct {
	APIKey string
	City   string
}

type Cond struct {
	Clouds          int     `json:"cloudcover,string"`
	FeelsLikeC      int     `json:",string"`
	Humidity        float64 `json:"humidity,string"`
	ObservationTime string  `json:"Observation_time"`
	PrecipMM        float64 `json:"precipMM,string"`
	Pressure        int     `json:"pressure,string"`
	TempC           int     `json:"temp_C,string"`
	VisibleDistKM   int     `json:"visibility,string"`
	WeatherCode     int     `json:"weatherCode,string"`
	WeatherIconUrl  []struct{ Value string }
	Winddir16Point  string
	WinddirDegree   int `json:"winddirDegree,string"`
	WindspeedKmph   int `json:"windspeedKmph,string"`
	WindspeedMiles  int `json:"windspeedMiles,string"`
}

type Astro struct {
	Moonrise string
	Moonset  string
	Sunrise  string
	Sunset   string
}

type Weather struct {
	Astronomy []Astro
	Date      string
	Hourly    []Cond
	MaxtempC  int `json:"maxtempC,string"`
	MintempC  int `json:"mintempC,string"`
}

type Resp struct {
	Data struct {
		Cur     []Cond     `json:"current_condition"`
		Req     []struct{} `json:"request"`
		Weather []Weather  `json:"weather"`
	} `json:"data"`
}

type State struct {
	Desc      string
	DayIcon   []string
	NightIcon []string
}

var (
	config     Config
	configpath string
	params     []string
	codes      = map[int]State{
		0:   {"Unknown, please report", iconUnknown, iconUnknown},
		113: {"Klar", iconSunny, iconUnknown},
		116: {"Teilweise Wolkig", iconPartlyCloudy, iconUnknown},
	}

	iconUnknown = []string{
		"    .-.      ",
		"     __)     ",
		"    (        ",
		"     `-’     ",
		"      •      "}
	iconSunny = []string{
		"\033[38;5;226m    \\   /    \033[0m",
		"\033[38;5;226m     .-.     \033[0m",
		"\033[38;5;226m  ― (   ) ―  \033[0m",
		"\033[38;5;226m     `-’     \033[0m",
		"\033[38;5;226m    /   \\    \033[0m"}
	iconPartlyCloudy = []string{
		"\033[38;5;226m   \\  /\033[0m      ",
		"\033[38;5;226m _ /\"\"\033[0m.-.    ",
		"\033[38;5;226m   \\_\033[0m(   ).  ",
		"\033[38;5;226m   /\033[0m(___(__) ",
		"             "}
	iconLightShowers = []string{
		"\033[38;5;226m _`/\"\"\033[38;5;250m.-.    \033[0m",
		"\033[38;5;226m  ,\\_\033[38;5;250m(   ).  \033[0m",
		"\033[38;5;226m   /\033[38;5;250m(___(__) \033[0m",
		"\033[38;5;111m     ‘ ‘ ‘ ‘ \033[0m",
		"\033[38;5;111m    ‘ ‘ ‘ ‘  \033[0m"}
	iconHeavyShowers = []string{
		"\033[38;5;226m _`/\"\"\033[38;5;240;1m.-.    \033[0m",
		"\033[38;5;226m  ,\\_\033[38;5;240;1m(   ).  \033[0m",
		"\033[38;5;226m   /\033[38;5;240;1m(___(__) \033[0m",
		"\033[38;5;21;1m   ‚‘‚‘‚‘‚‘ \033[0m",
		"\033[38;5;21;1m   ‚’‚’‚’‚’ \033[0m"}
	iconLightSnow = []string{
		"\033[38;5;226m _`/\"\"\033[38;5;250m.-.    \033[0m",
		"\033[38;5;226m  ,\\_\033[38;5;250m(   ).  \033[0m",
		"\033[38;5;226m   /\033[38;5;250m(___(__) \033[0m",
		"\033[38;5;255m     *  *  * \033[0m",
		"\033[38;5;255m    *  *  *  \033[0m"}
	iconHeavySnow = []string{
		"\033[38;5;226m _`/\"\"\033[38;5;240;1m.-.    \033[0m",
		"\033[38;5;226m  ,\\_\033[38;5;240;1m(   ).  \033[0m",
		"\033[38;5;226m   /\033[38;5;240;1m(___(__) \033[0m",
		"\033[38;5;255;1m    * * * *  \033[0m",
		"\033[38;5;255;1m   * * * *   \033[0m"}
	iconThunderyShowers = []string{
		"\033[38;5;226m _`/\"\"\033[38;5;250m.-.    \033[0m",
		"\033[38;5;226m  ,\\_\033[38;5;250m(   ).  \033[0m",
		"\033[38;5;226m   /\033[38;5;250m(___(__) \033[0m",
		"    \033[38;5;228;5m⚡\033[38;5;111;25m‘ ‘\033[38;5;228;5m⚡\033[38;5;111;25m‘ ‘ \033[0m",
		"    \033[38;5;111m‘ ‘ ‘ ‘  \033[0m"}
	iconThunderySnow = []string{
		"\033[38;5;226m _`/\"\"\033[38;5;250m.-.    \033[0m",
		"\033[38;5;226m  ,\\_\033[38;5;250m(   ).  \033[0m",
		"\033[38;5;226m   /\033[38;5;250m(___(__) \033[0m",
		"\033[38;5;255m     *\033[38;5;228;5m⚡\033[38;5;255;25m *\033[38;5;228;5m⚡\033[38;5;255;25m * \033[0m",
		"\033[38;5;255m    *  *  *  \033[0m"}

//	cloud = "
//   .-.
//  (   ).
// (___(__)
//  ' ' '
// ' ' ' "
)

const (
	uri = "https://api.worldweatheronline.com/free/v2/weather.ashx?"
)

func configload() error {
	if b, err := ioutil.ReadFile(configpath); err == nil {
		return json.Unmarshal(b, &config)
	} else {
		return err
	}
}

func configsave() error {
	if j, err := json.MarshalIndent(config, "", "\t"); err == nil {
		return ioutil.WriteFile(configpath, j, 0600)
	} else {
		return err
	}
}

func init() {
	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}
	configpath = path.Join(usr.HomeDir, ".wegorc")
	config.APIKey = ""
	config.City = "New York"
	err = configload()
	if _, ok := err.(*os.PathError); ok {
		if err := configsave(); err != nil {
			log.Fatal(err)
		}
	} else if err != nil {
		log.Fatalf("could not parse %v: %v", configpath, err)
	}
}

func main() {
	for i := 1; i < len(os.Args); i++ {
		_ = os.Args[i]
	}

	//	if len(config.APIKey) > 0 {
	//		params = append(params, "key="+config.APIKey)
	//	}
	//	if len(config.City) > 0 {
	//		params = append(params, "q="+url.QueryEscape(config.City))
	//	}
	//	params = append(params, "format=json")
	//	params = append(params, "num_of_days=3")
	//	params = append(params, "tp=3")
	//	params = append(params, "lang=de")
	//
	//	res, err := http.Get(uri + strings.Join(params, "&"))
	//	defer res.Body.Close()
	//	if err != nil {
	//		log.Fatal(err)
	//	}
	//	body, err := ioutil.ReadAll(res.Body)
	//	if err != nil {
	//		log.Fatal(err)
	//	}
	//
	//	//	fmt.Println(string(body))
	//
	//	var r Resp
	//	if err = json.Unmarshal(body, &r); err != nil {
	//		log.Println(err)
	//	}
	//
	//	temp := r.Data.Cur[0].FeelsLikeC
	//	fmt.Println(temp)
	fmt.Println(strings.Join(iconSunny, "|\n"))
	fmt.Println(strings.Join(iconPartlyCloudy, "|\n"))
	fmt.Println(strings.Join(iconHeavyShowers, "|\n"))
	fmt.Println(strings.Join(iconThunderyShowers, "|\n"))
}
