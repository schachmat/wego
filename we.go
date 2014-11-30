package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"time"
	"log"
	"net/http"
	"net/url"
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
	ChanceOfRain   string  `json:"chanceofrain"`
	FeelsLikeC     int     `json:",string"`
	PrecipMM       float32 `json:"precipMM,string"`
	TempC          int     `json:"tempC,string"`
	Time           string  `json:"time"`
	VisibleDistKM  int     `json:"visibility,string"`
	WeatherCode    int     `json:"weatherCode,string"`
	WeatherDesc    []struct{ Value string }
	WindGustKmph   int    `json:",string"`
	Winddir16Point string
	WindspeedKmph  int    `json:"windspeedKmph,string"`
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

type Icon struct {
	DayIcon   []string
	NightIcon []string
}

var (
	config     Config
	configpath string
	params     []string
	windDir    = map[string]string{
		"N":   "\033[1m↓\033[0m",
		"NNE": "\033[1m↓\033[0m",
		"NE":  "\033[1m↙\033[0m",
		"ENE": "\033[1m↙\033[0m",
		"E":   "\033[1m←\033[0m",
		"ESE": "\033[1m←\033[0m",
		"SE":  "\033[1m↖\033[0m",
		"SSE": "\033[1m↖\033[0m",
		"S":   "\033[1m↑\033[0m",
		"SSW": "\033[1m↑\033[0m",
		"SW":  "\033[1m↗\033[0m",
		"WSW": "\033[1m↗\033[0m",
		"W":   "\033[1m→\033[0m",
		"WNW": "\033[1m→\033[0m",
		"NW":  "\033[1m↘\033[0m",
		"NNW": "\033[1m↘\033[0m",
	}
	codes = map[int]Icon{
		0:   {iconUnknown, iconUnknown},
		113: {iconSunny, iconUnknown},
		116: {iconPartlyCloudy, iconUnknown},
		119: {iconUnknown, iconUnknown},
		122: {iconVeryCloudy, iconVeryCloudy},
		143: {iconUnknown, iconUnknown},
		176: {iconUnknown, iconUnknown},
		179: {iconUnknown, iconUnknown},
		182: {iconUnknown, iconUnknown},
		185: {iconUnknown, iconUnknown},
		200: {iconUnknown, iconUnknown},
		227: {iconUnknown, iconUnknown},
		230: {iconUnknown, iconUnknown},
		248: {iconUnknown, iconUnknown},
		260: {iconUnknown, iconUnknown},
		263: {iconUnknown, iconUnknown},
		266: {iconUnknown, iconUnknown},
		281: {iconUnknown, iconUnknown},
		284: {iconUnknown, iconUnknown},
		293: {iconUnknown, iconUnknown},
		296: {iconUnknown, iconUnknown},
		299: {iconUnknown, iconUnknown},
		302: {iconUnknown, iconUnknown},
		305: {iconUnknown, iconUnknown},
		308: {iconUnknown, iconUnknown},
		311: {iconUnknown, iconUnknown},
		314: {iconUnknown, iconUnknown},
		317: {iconUnknown, iconUnknown},
		320: {iconUnknown, iconUnknown},
		323: {iconUnknown, iconUnknown},
		326: {iconUnknown, iconUnknown},
		329: {iconUnknown, iconUnknown},
		332: {iconUnknown, iconUnknown},
		335: {iconUnknown, iconUnknown},
		338: {iconUnknown, iconUnknown},
		350: {iconUnknown, iconUnknown},
		353: {iconUnknown, iconUnknown},
		356: {iconUnknown, iconUnknown},
		359: {iconUnknown, iconUnknown},
		362: {iconUnknown, iconUnknown},
		365: {iconUnknown, iconUnknown},
		368: {iconUnknown, iconUnknown},
		371: {iconUnknown, iconUnknown},
		374: {iconUnknown, iconUnknown},
		377: {iconUnknown, iconUnknown},
		386: {iconUnknown, iconUnknown},
		389: {iconUnknown, iconUnknown},
		392: {iconUnknown, iconUnknown},
		395: {iconUnknown, iconUnknown},
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
	iconVeryCloudy = []string{
		"             ",
		"\033[38;5;240m     .--.    \033[0m",
		"\033[38;5;240m  .-(    ).  \033[0m",
		"\033[38;5;240m (___.__)__) \033[0m",
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

func formatTemp(c Cond) string {
	return fmt.Sprintf("%d – %d °C", c.FeelsLikeC, c.TempC)
}

func formatWind(c Cond) string {
	if c.WindGustKmph > c.WindspeedKmph {
		return fmt.Sprintf("%s %d – %d km/h", windDir[c.Winddir16Point], c.WindspeedKmph, c.WindGustKmph)
	} else {
		return fmt.Sprintf("%s %d km/h", windDir[c.Winddir16Point], c.WindspeedKmph)
	}
}

func formatVisibility(c Cond) string {
	return fmt.Sprintf("%d km", c.VisibleDistKM)
}

func formatRain(c Cond) string {
	if c.ChanceOfRain != "" {
		return fmt.Sprintf("%v mm | %s%%", c.PrecipMM, c.ChanceOfRain)
	} else {
		return fmt.Sprintf("%v mm", c.PrecipMM)
	}
}

func formatCond(cur []string, c Cond) (ret []string) {
	var icon Icon
	if i, ok := codes[c.WeatherCode]; !ok {
		i.DayIcon = iconUnknown
		i.NightIcon = iconUnknown
		icon = i
	} else {
		icon = i
	}
	ret = append(ret, fmt.Sprintf("%v %v %-15v", cur[0], icon.DayIcon[0], c.WeatherDesc[0].Value))
	ret = append(ret, fmt.Sprintf("%v %v %-15v", cur[1], icon.DayIcon[1], formatTemp(c)))
	ret = append(ret, fmt.Sprintf("%v %v %-23v", cur[2], icon.DayIcon[2], formatWind(c)))
	ret = append(ret, fmt.Sprintf("%v %v %-15v", cur[3], icon.DayIcon[3], formatVisibility(c)))
	ret = append(ret, fmt.Sprintf("%v %v %-15v", cur[4], icon.DayIcon[4], formatRain(c)))
	return
}

func printDay(w Weather) (ret []string) {
	hourly := w.Hourly
	ret = make([]string, 5)
	for i := range ret {
		ret[i] = "│"
	}
	for _, h := range hourly {
		if h.Time == "100" || h.Time == "400" || h.Time == "700" || h.Time == "1600" {
			continue
		}
		ret = formatCond(ret, h)
		for i := range ret {
			ret[i] = ret[i] + "│"
		}
	}
	d, _ := time.Parse("2006-01-02", w.Date)
	dateFmt := "┤ " + d.Format("Mon 02. Jan") + " ├"
	ret = append([]string{
		"                                                       ┌─────────────┐                                                       ",
		"┌──────────────────────────────┬───────────────────────" + dateFmt + "───────────────────────┬──────────────────────────────┐",
		"│           Morning            │             Noon      └──────┬──────┘    Evening            │            Night             │",
		"├──────────────────────────────┼──────────────────────────────┼──────────────────────────────┼──────────────────────────────┤"},
		ret...)
	return append(ret,
		"└──────────────────────────────┴──────────────────────────────┴──────────────────────────────┴──────────────────────────────┘")
	return
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

	if len(config.APIKey) > 0 {
		params = append(params, "key="+config.APIKey)
	}
	if len(config.City) > 0 {
		params = append(params, "q="+url.QueryEscape(config.City))
	}
	params = append(params, "format=json")
	params = append(params, "num_of_days=3")
	params = append(params, "tp=3")
	params = append(params, "lang=de")

	res, err := http.Get(uri + strings.Join(params, "&"))
	defer res.Body.Close()
	if err != nil {
		log.Fatal(err)
	}
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err)
	}

//	fmt.Println(string(body))

	var r Resp
	if err = json.Unmarshal(body, &r); err != nil {
		log.Println(err)
	}

	out := formatCond(make([]string, 5), r.Data.Cur[0])
	for _, val := range out {
		fmt.Println(val)
	}

	for _, d := range r.Data.Weather {
		for _, val := range printDay(d) {
			fmt.Println(val)
		}
	}
	//	fmt.Println(strings.Join(iconSunny, "|\n"))
	//	fmt.Println(strings.Join(iconPartlyCloudy, "|\n"))
	//	fmt.Println(strings.Join(iconHeavyShowers, "|\n"))
	//	fmt.Println(strings.Join(iconThunderyShowers, "|\n"))
}
