package frontends

import (
	"flag"
	"fmt"
	"log"
	"math"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/mattn/go-colorable"
	"github.com/mattn/go-runewidth"
	"github.com/schachmat/wego/iface"
)

type asciiarttable struct {
	aatImperial bool
}

var (
	ansiEsc   *regexp.Regexp
	slotTimes = [slotcount]int{9 * 60, 12 * 60, 18 * 60, 22 * 60}
	windDir   = map[string]string{
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
	unitRain = map[bool]string{
		false: "mm",
		true:  "in",
	}
	unitTemp = map[bool]string{
		false: "C",
		true:  "F",
	}
	unitVis = map[bool]string{
		false: "km",
		true:  "mi",
	}
	unitWind = map[bool]string{
		false: "km/h",
		true:  "mph",
	}
	codes = map[int][]string{
		113: iconSunny,
		116: iconPartlyCloudy,
		119: iconCloudy,
		122: iconVeryCloudy,
		143: iconFog,
		176: iconLightShowers,
		179: iconLightSleetShowers,
		182: iconLightSleet,
		185: iconLightSleet,
		200: iconThunderyShowers,
		227: iconLightSnow,
		230: iconHeavySnow,
		248: iconFog,
		260: iconFog,
		263: iconLightShowers,
		266: iconLightRain,
		281: iconLightSleet,
		284: iconLightSleet,
		293: iconLightRain,
		296: iconLightRain,
		299: iconHeavyShowers,
		302: iconHeavyRain,
		305: iconHeavyShowers,
		308: iconHeavyRain,
		311: iconLightSleet,
		314: iconLightSleet,
		317: iconLightSleet,
		320: iconLightSnow,
		323: iconLightSnowShowers,
		326: iconLightSnowShowers,
		329: iconHeavySnow,
		332: iconHeavySnow,
		335: iconHeavySnowShowers,
		338: iconHeavySnow,
		350: iconLightSleet,
		353: iconLightShowers,
		356: iconHeavyShowers,
		359: iconHeavyRain,
		362: iconLightSleetShowers,
		365: iconLightSleetShowers,
		368: iconLightSnowShowers,
		371: iconHeavySnowShowers,
		374: iconLightSleetShowers,
		377: iconLightSleet,
		386: iconThunderyShowers,
		389: iconThunderyHeavyRain,
		392: iconThunderySnowShowers,
		395: iconHeavySnowShowers,
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
		"\033[38;5;21;1m   ‚‘‚‘‚‘‚‘  \033[0m",
		"\033[38;5;21;1m   ‚’‚’‚’‚’  \033[0m"}
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
	iconLightSleetShowers = []string{
		"\033[38;5;226m _`/\"\"\033[38;5;250m.-.    \033[0m",
		"\033[38;5;226m  ,\\_\033[38;5;250m(   ).  \033[0m",
		"\033[38;5;226m   /\033[38;5;250m(___(__) \033[0m",
		"\033[38;5;111m     ‘ \033[38;5;255m*\033[38;5;111m ‘ \033[38;5;255m* \033[0m",
		"\033[38;5;255m    *\033[38;5;111m ‘ \033[38;5;255m*\033[38;5;111m ‘  \033[0m"}
	iconThunderyShowers = []string{
		"\033[38;5;226m _`/\"\"\033[38;5;250m.-.    \033[0m",
		"\033[38;5;226m  ,\\_\033[38;5;250m(   ).  \033[0m",
		"\033[38;5;226m   /\033[38;5;250m(___(__) \033[0m",
		"\033[38;5;228;5m    ⚡\033[38;5;111;25m‘ ‘\033[38;5;228;5m⚡\033[38;5;111;25m‘ ‘ \033[0m",
		"\033[38;5;111m    ‘ ‘ ‘ ‘  \033[0m"}
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
	iconLightSleet = []string{
		"\033[38;5;250m     .-.     \033[0m",
		"\033[38;5;250m    (   ).   \033[0m",
		"\033[38;5;250m   (___(__)  \033[0m",
		"\033[38;5;111m    ‘ \033[38;5;255m*\033[38;5;111m ‘ \033[38;5;255m*  \033[0m",
		"\033[38;5;255m   *\033[38;5;111m ‘ \033[38;5;255m*\033[38;5;111m ‘   \033[0m"}
	iconFog = []string{
		"             ",
		"\033[38;5;251m _ - _ - _ - \033[0m",
		"\033[38;5;251m  _ - _ - _  \033[0m",
		"\033[38;5;251m _ - _ - _ - \033[0m",
		"             "}
)

const (
	slotcount = 4
)

func pad(s string, mustLen int) (ret string) {
	ret = s
	realLen := utf8.RuneCountInString(ansiEsc.ReplaceAllLiteralString(s, ""))
	delta := mustLen - realLen
	if delta > 0 {
		ret += "\033[0m" + strings.Repeat(" ", delta)
	} else if delta < 0 {
		toks := ansiEsc.Split(s, 2)
		tokLen := utf8.RuneCountInString(toks[0])
		esc := ansiEsc.FindString(s)
		if tokLen > mustLen {
			ret = fmt.Sprintf("%.*s\033[0m", mustLen, toks[0])
		} else {
			ret = fmt.Sprintf("%s%s%s", toks[0], esc, pad(toks[1], mustLen-tokLen))
		}
	}
	return
}

func (c *asciiarttable) formatTemp(cond iface.Cond) string {
	color := func(temp int) string {
		var col = 21
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
			}
		}
		if c.aatImperial {
			temp = (temp*18 + 320) / 10
		}
		return fmt.Sprintf("\033[38;5;%03dm%d\033[0m", col, temp)
	}
	t := cond.TempC
	if t == 0 {
		t = cond.TempC2
	}
	if cond.FeelsLikeC < t {
		return pad(fmt.Sprintf("%s – %s °%s", color(cond.FeelsLikeC), color(t), unitTemp[c.aatImperial]), 15)
	} else if cond.FeelsLikeC > t {
		return pad(fmt.Sprintf("%s – %s °%s", color(t), color(cond.FeelsLikeC), unitTemp[c.aatImperial]), 15)
	}
	return pad(fmt.Sprintf("%s °%s", color(cond.FeelsLikeC), unitTemp[c.aatImperial]), 15)
}

func (c *asciiarttable) formatWind(cond iface.Cond) string {
	color := func(spd int) string {
		var col = 46
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
			}
		}
		if c.aatImperial {
			spd = (spd * 1000) / 1609
		}
		return fmt.Sprintf("\033[38;5;%03dm%d\033[0m", col, spd)
	}
	if cond.WindGustKmph > cond.WindspeedKmph {
		return pad(fmt.Sprintf("%s %s – %s %s", windDir[cond.Winddir16Point], color(cond.WindspeedKmph), color(cond.WindGustKmph), unitWind[c.aatImperial]), 15)
	}
	return pad(fmt.Sprintf("%s %s %s", windDir[cond.Winddir16Point], color(cond.WindspeedKmph), unitWind[c.aatImperial]), 15)
}

func (c *asciiarttable) formatVisibility(cond iface.Cond) string {
	if c.aatImperial {
		cond.VisibleDistKM = (cond.VisibleDistKM * 621) / 1000
	}
	return pad(fmt.Sprintf("%d %s", cond.VisibleDistKM, unitVis[c.aatImperial]), 15)
}

func (c *asciiarttable) formatRain(cond iface.Cond) string {
	rainUnit := float32(cond.PrecipMM)
	if c.aatImperial {
		rainUnit = float32(cond.PrecipMM) * 0.039
	}
	if cond.ChanceOfRain != "" {
		return pad(fmt.Sprintf("%.1f %s | %s%%", rainUnit, unitRain[c.aatImperial], cond.ChanceOfRain), 15)
	}
	return pad(fmt.Sprintf("%.1f %s", rainUnit, unitRain[c.aatImperial]), 15)
}

func (c *asciiarttable) formatCond(cur []string, cond iface.Cond, current bool) (ret []string) {
	var icon []string
	if i, ok := codes[cond.WeatherCode]; !ok {
		icon = iconUnknown
	} else {
		icon = i
	}
	desc := cond.WeatherDesc[0].Value
	if !current {
		desc = runewidth.Truncate(runewidth.FillRight(desc, 15), 15, "…")
	}
	ret = append(ret, fmt.Sprintf("%v %v %v", cur[0], icon[0], desc))
	ret = append(ret, fmt.Sprintf("%v %v %v", cur[1], icon[1], c.formatTemp(cond)))
	ret = append(ret, fmt.Sprintf("%v %v %v", cur[2], icon[2], c.formatWind(cond)))
	ret = append(ret, fmt.Sprintf("%v %v %v", cur[3], icon[3], c.formatVisibility(cond)))
	ret = append(ret, fmt.Sprintf("%v %v %v", cur[4], icon[4], c.formatRain(cond)))
	return
}

func (c *asciiarttable) printDay(w iface.Weather) (ret []string) {
	hourly := w.Hourly
	ret = make([]string, 5)
	for i := range ret {
		ret[i] = "│"
	}

	// find hourly data which fits the desired times of day best
	var slots [slotcount]iface.Cond
	for _, h := range hourly {
		cur := int(math.Mod(float64(h.Time), 100)) + 60*(h.Time/100)
		for i, s := range slots {
			if math.Abs(float64(cur-slotTimes[i])) < math.Abs(float64(s.Time-slotTimes[i])) {
				h.Time = cur
				slots[i] = h
			}
		}
	}

	for _, s := range slots {
		ret = c.formatCond(ret, s, false)
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
}

func (c *asciiarttable) Setup() {
	flag.BoolVar(&c.aatImperial, "aat-imperial", false, "use imperial units for output")
}

func (c *asciiarttable) Render(r iface.Resp) {
	fmt.Printf("Weather for %s: %s\n\n", r.Data.Req[0].Type, r.Data.Req[0].Query)
	stdout := colorable.NewColorableStdout()

	if r.Data.Cur == nil || len(r.Data.Cur) < 1 {
		log.Fatal("No weather data available.")
	}
	out := c.formatCond(make([]string, 5), r.Data.Cur[0], true)
	for _, val := range out {
		fmt.Fprintln(stdout, val)
	}

	if len(r.Data.Weather) == 0 {
		return
	}
	if r.Data.Weather == nil {
		log.Fatal("No detailed weather forecast available.")
	}
	for _, d := range r.Data.Weather {
		for _, val := range c.printDay(d) {
			fmt.Fprintln(stdout, val)
		}
	}
}

func init() {
	ansiEsc = regexp.MustCompile("\033.*?m")
	All["ascii-art-table"] = &asciiarttable{}
}
