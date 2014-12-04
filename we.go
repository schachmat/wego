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
	codes = map[int][]string{
		113: iconSunny,
		116: iconPartlyCloudy,
		119: iconCloudy,
		122: iconVeryCloudy,
		143: iconUnknown,
		176: iconUnknown,
		179: iconUnknown,
		182: iconUnknown,
		185: iconUnknown,
		200: iconThunderyShowers,
		227: iconLightSnow,
		230: iconHeavySnow,
		248: iconUnknown,
		260: iconUnknown,
		263: iconUnknown,
		266: iconLightRain,
		281: iconUnknown,
		284: iconUnknown,
		293: iconLightRain,
		296: iconLightRain,
		299: iconUnknown,
		302: iconHeavyRain,
		305: iconUnknown,
		308: iconHeavyRain,
		311: iconUnknown,
		314: iconUnknown,
		317: iconUnknown,
		320: iconLightSnow,
		323: iconLightSnowShowers,
		326: iconLightSnowShowers,
		329: iconHeavySnow,
		332: iconHeavySnow,
		335: iconHeavySnowShowers,
		338: iconHeavySnow,
		350: iconUnknown,
		353: iconUnknown,
		356: iconUnknown,
		359: iconHeavyRain,
		362: iconUnknown,
		365: iconUnknown,
		368: iconLightSnowShowers,
		371: iconHeavySnowShowers,
		374: iconUnknown,
		377: iconUnknown,
		386: iconThunderyShowers,
		389: iconThunderyHeavyRain,
		392: iconThunderySnowShowers,
		395: iconHeavySnowShowers, // ThunderyHeavySnow
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
		"\033[38;5;226m _ /\"\"\033[38;5;250m.-.    \033[0m",
		"\033[38;5;226m   \\_\033[38;5;250m(   ).  \033[0m",
		"\033[38;5;226m   /\033[38;5;250m(___(__) \033[0m",
		"             "}
	iconCloudy = []string{
		"             ",
		"\033[38;5;250m     .--.    \033[0m",
		"\033[38;5;250m  .-(    ).  \033[0m",
		"\033[38;5;250m (___.__)__) \033[0m",
		"             "}
	iconVeryCloudy = []string{
		"             ",
		"\033[38;5;240;1m     .--.    \033[0m",
		"\033[38;5;240;1m  .-(    ).  \033[0m",
		"\033[38;5;240;1m (___.__)__) \033[0m",
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
	iconLightSnowShowers = []string{
		"\033[38;5;226m _`/\"\"\033[38;5;250m.-.    \033[0m",
		"\033[38;5;226m  ,\\_\033[38;5;250m(   ).  \033[0m",
		"\033[38;5;226m   /\033[38;5;250m(___(__) \033[0m",
		"\033[38;5;255m     *  *  * \033[0m",
		"\033[38;5;255m    *  *  *  \033[0m"}
	iconHeavySnowShowers = []string{
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
	iconThunderyHeavyRain = []string{
		"\033[38;5;240;1m     .-.     \033[0m",
		"\033[38;5;240;1m    (   ).   \033[0m",
		"\033[38;5;240;1m   (___(__)  \033[0m",
		"\033[38;5;21;1m  ‚‘\033[38;5;228;5m⚡\033[38;5;21;25m‘‚\033[38;5;228;5m⚡\033[38;5;21;25m‚‘   \033[0m",
		"\033[38;5;21;1m  ‚’‚’\033[38;5;228;5m⚡\033[38;5;21;25m’‚’   \033[0m"}
	iconThunderySnowShowers = []string{
		"\033[38;5;226m _`/\"\"\033[38;5;250m.-.    \033[0m",
		"\033[38;5;226m  ,\\_\033[38;5;250m(   ).  \033[0m",
		"\033[38;5;226m   /\033[38;5;250m(___(__) \033[0m",
		"\033[38;5;255m     *\033[38;5;228;5m⚡\033[38;5;255;25m *\033[38;5;228;5m⚡\033[38;5;255;25m * \033[0m",
		"\033[38;5;255m    *  *  *  \033[0m"}
	iconLightRain = []string{
		"\033[38;5;250m     .-.     \033[0m",
		"\033[38;5;250m    (   ).   \033[0m",
		"\033[38;5;250m   (___(__)  \033[0m",
		"\033[38;5;111m    ‘ ‘ ‘ ‘  \033[0m",
		"\033[38;5;111m   ‘ ‘ ‘ ‘   \033[0m"}
	iconHeavyRain = []string{
		"\033[38;5;240;1m     .-.     \033[0m",
		"\033[38;5;240;1m    (   ).   \033[0m",
		"\033[38;5;240;1m   (___(__)  \033[0m",
		"\033[38;5;21;1m  ‚‘‚‘‚‘‚‘   \033[0m",
		"\033[38;5;21;1m  ‚’‚’‚’‚’   \033[0m"}
	iconLightSnow = []string{
		"\033[38;5;250m     .-.     \033[0m",
		"\033[38;5;250m    (   ).   \033[0m",
		"\033[38;5;250m   (___(__)  \033[0m",
		"\033[38;5;255m    *  *  *  \033[0m",
		"\033[38;5;255m   *  *  *   \033[0m"}
	iconHeavySnow = []string{
		"\033[38;5;240;1m     .-.     \033[0m",
		"\033[38;5;240;1m    (   ).   \033[0m",
		"\033[38;5;240;1m   (___(__)  \033[0m",
		"\033[38;5;255;1m   * * * *   \033[0m",
		"\033[38;5;255;1m  * * * *    \033[0m"}

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
	color := func(temp int) string {
		var col int = 21
		switch temp {
		case -15, -14, -13: col = 27
		case -12, -11, -10: col = 33
		case -9, -8, -7: col = 39
		case -6, -5, -4: col = 45
		case -3, -2, -1: col = 51
		case 0, 1: col = 50
		case 2, 3: col = 49
		case 4, 5: col = 48
		case 6, 7: col = 47
		case 8, 9: col = 46
		case 10, 11, 12: col = 82
		case 13, 14, 15: col = 118
		case 16, 17, 18: col = 154
		case 19, 20, 21: col = 190
		case 22, 23, 24: col = 226
		case 25, 26, 27: col = 220
		case 28, 29, 30: col = 214
		case 31, 32, 33: col = 208
		case 34, 35, 36: col = 202
		default:
			if temp > 0 {
				col = 196
			}
		}
		return fmt.Sprintf("\033[38;5;%03dm%d\033[0m", col, temp)
	}
	if c.FeelsLikeC != c.TempC {
		return fmt.Sprintf("%s – %s °C         ", color(c.FeelsLikeC), color(c.TempC))[:48]
	} else {
		return fmt.Sprintf("%s °C            ", color(c.FeelsLikeC))[:31]
	}
}

func formatWind(c Cond) string {
	color := func(spd int) string {
		var col int = 46
		switch spd {
		case 1, 2, 3: col = 82
		case 4, 5, 6: col = 118
		case 7, 8, 9: col = 154
		case 10, 11, 12: col = 190
		case 13, 14, 15: col = 226
		case 16, 17, 18, 19: col = 220
		case 20, 21, 22, 23: col = 214
		case 24, 25, 26, 27: col = 208
		case 28, 29, 30, 31: col = 202
		default:
			if spd > 0 {
				col = 196
			}
		}
		return fmt.Sprintf("\033[38;5;%03dm%d\033[0m", col, spd)
	}
	if c.WindGustKmph > c.WindspeedKmph {
		return fmt.Sprintf("%s %s – %s km/h     ", windDir[c.Winddir16Point], color(c.WindspeedKmph), color(c.WindGustKmph))[:57]
	} else {
		return fmt.Sprintf("%s %s km/h       ", windDir[c.Winddir16Point], color(c.WindspeedKmph))[:40]
	}
}

func formatVisibility(c Cond) string {
	return fmt.Sprintf("%d km            ", c.VisibleDistKM)[:15]
}

func formatRain(c Cond) string {
	if c.ChanceOfRain != "" {
		return fmt.Sprintf("%v mm | %s%%        ", c.PrecipMM, c.ChanceOfRain)[:15]
	} else {
		return fmt.Sprintf("%v mm            ", c.PrecipMM)[:15]
	}
}

func formatCond(cur []string, c Cond) (ret []string) {
	var icon []string
	if i, ok := codes[c.WeatherCode]; !ok {
		icon = iconUnknown
	} else {
		icon = i
	}
	ret = append(ret, fmt.Sprintf("%v %v %-15v", cur[0], icon[0], c.WeatherDesc[0].Value))
	ret = append(ret, fmt.Sprintf("%v %v %v", cur[1], icon[1], formatTemp(c)))
	ret = append(ret, fmt.Sprintf("%v %v %v", cur[2], icon[2], formatWind(c)))
	ret = append(ret, fmt.Sprintf("%v %v %v", cur[3], icon[3], formatVisibility(c)))
	ret = append(ret, fmt.Sprintf("%v %v %v", cur[4], icon[4], formatRain(c)))
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
}
