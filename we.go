package main

import (
	"encoding/json"
	"fmt"
	"github.com/mattn/go-colorable"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/user"
	"path"
	"strconv"
	"strings"
	"time"
)

const (
	uri = "https://api.worldweatheronline.com/free/v2/weather.ashx?"
)

func configload() error {
	b, err := ioutil.ReadFile(configpath)
	if err == nil {
		return json.Unmarshal(b, &config)
	}
	return err
}

func configsave() error {
	j, err := json.MarshalIndent(config, "", "\t")
	if err == nil {
		return ioutil.WriteFile(configpath, j, 0600)
	}
	return err
}

func tempColor(temp int) string {
	var col int
	switch temp {
	case -15, -14, -13:
		col = 27
	case -12, -11, -10:
		col = 33
	case -9, -8, -7:
		col = 39
	case -6, -5, -4:
		col = 45
	case -3, -2, -1:
		col = 51
	case 0, 1:
		col = 50
	case 2, 3:
		col = 49
	case 4, 5:
		col = 48
	case 6, 7:
		col = 47
	case 8, 9:
		col = 46
	case 10, 11, 12:
		col = 82
	case 13, 14, 15:
		col = 118
	case 16, 17, 18:
		col = 154
	case 19, 20, 21:
		col = 190
	case 22, 23, 24:
		col = 226
	case 25, 26, 27:
		col = 220
	case 28, 29, 30:
		col = 214
	case 31, 32, 33:
		col = 208
	case 34, 35, 36:
		col = 202
	default:
		if temp > 0 {
			col = 196
		} else {
			col = 21
		}
	}
	tempUnit := float32(temp)
	if config.Imperial {
		tempUnit = float32(temp)*1.8 + 32.0
	}
	return fmt.Sprintf("\033[38;5;%03dm%d\033[0m", col, int32(tempUnit))
}

func formatTemp(c cond) string {
	if c.FeelsLikeC < c.TempC {
		return fmt.Sprintf("%s – %s °%s         ", tempColor(c.FeelsLikeC), tempColor(c.TempC), unitTemp[config.Imperial])[:48]
	} else if c.FeelsLikeC > c.TempC {
		return fmt.Sprintf("%s – %s °%s         ", tempColor(c.TempC), tempColor(c.FeelsLikeC), unitTemp[config.Imperial])[:48]
	}
	return fmt.Sprintf("%s °%s            ", tempColor(c.FeelsLikeC), unitTemp[config.Imperial])[:31]
}

func windColor(spd int) string {
	var col int
	switch spd {
	case 1, 2, 3:
		col = 82
	case 4, 5, 6:
		col = 118
	case 7, 8, 9:
		col = 154
	case 10, 11, 12:
		col = 190
	case 13, 14, 15:
		col = 226
	case 16, 17, 18, 19:
		col = 220
	case 20, 21, 22, 23:
		col = 214
	case 24, 25, 26, 27:
		col = 208
	case 28, 29, 30, 31:
		col = 202
	default:
		if spd > 0 {
			col = 196
		} else {
			col = 46
		}
	}
	spdUnit := float32(spd)
	if config.Imperial {
		spdUnit = float32(spd) / 1.609
	}
	return fmt.Sprintf("\033[38;5;%03dm%d\033[0m", col, int32(spdUnit))
}

func formatWind(c cond) string {
	if c.WindGustKmph > c.WindspeedKmph {
		return fmt.Sprintf("%s %s – %s %s     ", windDir[c.Winddir16Point], windColor(c.WindspeedKmph), windColor(c.WindGustKmph), unitWind[config.Imperial])[:57]
	}
	return fmt.Sprintf("%s %s %s        ", windDir[c.Winddir16Point], windColor(c.WindspeedKmph), unitWind[config.Imperial])[:40]
}

func formatVisibility(c cond) string {
	distUnit := float32(c.VisibleDistKM)
	if config.Imperial {
		distUnit = float32(c.VisibleDistKM) * 0.621
	}
	return fmt.Sprintf("%d %s            ", int32(distUnit), unitVis[config.Imperial])[:15]
}

func formatRain(c cond) string {
	rainUnit := float32(c.PrecipMM)
	if config.Imperial {
		rainUnit = float32(c.PrecipMM) * 0.039
	}
	if c.ChanceOfRain != "" {
		return fmt.Sprintf("%.1f %s | %s%%        ", rainUnit, unitRain[config.Imperial], c.ChanceOfRain)[:15]
	}
	return fmt.Sprintf("%.1f %s            ", rainUnit, unitRain[config.Imperial])[:15]
}

func formatCond(cur []string, c cond) (ret []string) {
	var icon []string
	if i, ok := codes[c.WeatherCode]; !ok {
		icon = iconUnknown
	} else {
		icon = i
	}
	ret = append(ret, fmt.Sprintf("%v %v %-15.15v", cur[0], icon[0], c.WeatherDesc[0].Value))
	ret = append(ret, fmt.Sprintf("%v %v %v", cur[1], icon[1], formatTemp(c)))
	ret = append(ret, fmt.Sprintf("%v %v %v", cur[2], icon[2], formatWind(c)))
	ret = append(ret, fmt.Sprintf("%v %v %v", cur[3], icon[3], formatVisibility(c)))
	ret = append(ret, fmt.Sprintf("%v %v %v", cur[4], icon[4], formatRain(c)))
	return
}

func printDay(w weather) (ret []string) {
	hourly := w.Hourly
	ret = make([]string, 5)
	for i := range ret {
		ret[i] = "│"
	}
	for _, h := range hourly {
		if h.Time == "0" || h.Time == "100" ||
			h.Time == "200" || h.Time == "300" || h.Time == "400" ||
			h.Time == "500" || h.Time == "600" || h.Time == "700" ||
			h.Time == "1400" || h.Time == "1500" || h.Time == "1600" ||
			h.Time == "2300" {
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
	config.Imperial = false
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
	var numdays = 3

	if len(config.APIKey) == 0 {
		log.Fatal("No API key specified. Setup instructions are in the README.")
	}
	params = append(params, "key="+config.APIKey)

	for _, arg := range os.Args[1:] {
		if v, err := strconv.Atoi(arg); err == nil {
			numdays = v
		} else {
			config.City = arg
		}
	}

	if len(config.City) > 0 {
		params = append(params, "q="+url.QueryEscape(config.City))
	}
	params = append(params, "format=json")
	params = append(params, "num_of_days="+strconv.Itoa(numdays))
	params = append(params, "tp=3")
	params = append(params, "lang=de")

	// fmt.Fprintln(os.Stderr, params)

	res, err := http.Get(uri + strings.Join(params, "&"))
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	// fmt.Println(string(body))

	var r resp
	if err = json.Unmarshal(body, &r); err != nil {
		log.Println(err)
	}

	if r.Data.Req == nil || len(r.Data.Req) < 1 {
		if r.Data.Err != nil && len(r.Data.Err) >= 1 {
			log.Fatal(r.Data.Err[0].Msg)
		}
		log.Fatal("Malformed response.")
	}
	fmt.Printf("Weather for %s: %s\n\n", r.Data.Req[0].Type, r.Data.Req[0].Query)
	stdout := colorable.NewColorableStdout()

	if r.Data.Cur == nil || len(r.Data.Cur) < 1 {
		log.Fatal("No weather data available.")
	}
	out := formatCond(make([]string, 5), r.Data.Cur[0])
	for _, val := range out {
		fmt.Fprintln(stdout, val)
	}

	if numdays == 0 {
		return
	}
	if r.Data.Weather == nil {
		log.Fatal("No detailed weather forecast available.")
	}
	for _, d := range r.Data.Weather {
		for _, val := range printDay(d) {
			fmt.Fprintln(stdout, val)
		}
	}
}
