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

type aatConfig struct {
	imperial bool
}

//TODO: replace s parameter with printf interface?
func aatPad(s string, mustLen int) (ret string) {
	ansiEsc := regexp.MustCompile("\033.*?m")
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
			ret = fmt.Sprintf("%s%s%s", toks[0], esc, aatPad(toks[1], mustLen-tokLen))
		}
	}
	return
}

func (c *aatConfig) formatTemp(cond iface.Cond) string {
	unit := map[bool]string{
		false: "C",
		true:  "F",
	}
	color := func(temp float32) string {
		col := 196
		if temp < -15 {
			col = 21
		} else if temp < -12 {
			col = 27
		} else if temp < -9 {
			col = 33
		} else if temp < -6 {
			col = 39
		} else if temp < -3 {
			col = 45
		} else if temp < 0 {
			col = 51
		} else if temp < 2 {
			col = 50
		} else if temp < 4 {
			col = 49
		} else if temp < 6 {
			col = 48
		} else if temp < 8 {
			col = 47
		} else if temp < 10 {
			col = 46
		} else if temp < 13 {
			col = 82
		} else if temp < 16 {
			col = 118
		} else if temp < 19 {
			col = 154
		} else if temp < 22 {
			col = 190
		} else if temp < 25 {
			col = 226
		} else if temp < 28 {
			col = 220
		} else if temp < 31 {
			col = 214
		} else if temp < 34 {
			col = 208
		} else if temp < 37 {
			col = 202
		}
		if c.imperial {
			temp = (temp*18 + 320) / 10
		}
		return fmt.Sprintf("\033[38;5;%03dm%d\033[0m", col, int(temp))
	}

	if cond.TempC == nil {
		return aatPad(fmt.Sprintf("? °%s", unit[c.imperial]), 15)
	}

	t := *cond.TempC
	if cond.FeelsLikeC != nil {
		fl := *cond.FeelsLikeC
		if fl < t {
			return aatPad(fmt.Sprintf("%s – %s °%s", color(fl), color(t), unit[c.imperial]), 15)
		} else if fl > t {
			return aatPad(fmt.Sprintf("%s – %s °%s", color(t), color(fl), unit[c.imperial]), 15)
		}
	}
	return aatPad(fmt.Sprintf("%s °%s", color(t), unit[c.imperial]), 15)
}

func (c *aatConfig) formatWind(cond iface.Cond) string {
	unit := map[bool]string{
		false: "km/h",
		true:  "mph",
	}
	windDir := func(deg *int) string {
		if deg == nil {
			return "?"
		}
		arrows := []string{"↓", "↙", "←", "↖", "↑", "↗", "→", "↘"}
		return "\033[1m" + arrows[((*deg+22)%360)/45] + "\033[0m"
	}
	color := func(spdKmph float32) string {
		col := 196
		if spdKmph <= 0 {
			col = 46
		} else if spdKmph < 4 {
			col = 82
		} else if spdKmph < 7 {
			col = 118
		} else if spdKmph < 10 {
			col = 154
		} else if spdKmph < 13 {
			col = 190
		} else if spdKmph < 16 {
			col = 226
		} else if spdKmph < 20 {
			col = 220
		} else if spdKmph < 24 {
			col = 214
		} else if spdKmph < 28 {
			col = 208
		} else if spdKmph < 32 {
			col = 202
		}
		if c.imperial {
			spdKmph = (spdKmph * 1000) / 1609
		}
		return fmt.Sprintf("\033[38;5;%03dm%d\033[0m", col, int(spdKmph))
	}

	if cond.WindspeedKmph == nil {
		return aatPad(windDir(cond.WinddirDegree), 15)
	}
	s := *cond.WindspeedKmph

	if cond.WindGustKmph != nil {
		if g := *cond.WindGustKmph; g > s {
			return aatPad(fmt.Sprintf("%s %s – %s %s", windDir(cond.WinddirDegree), color(s), color(g), unit[c.imperial]), 15)
		}
	}

	return aatPad(fmt.Sprintf("%s %s %s", windDir(cond.WinddirDegree), color(s), unit[c.imperial]), 15)
}

func (c *aatConfig) formatVisibility(cond iface.Cond) string {
	unit := map[bool]string{
		false: "km",
		true:  "mi",
	}
	if cond.VisibleDistKM == nil {
		return aatPad("", 15)
	}
	v := *cond.VisibleDistKM

	if c.imperial {
		v = (v * 621) / 1000
	}
	return aatPad(fmt.Sprintf("%d %s", int(v), unit[c.imperial]), 15)
}

func (c *aatConfig) formatRain(cond iface.Cond) string {
	unit := map[bool]string{
		false: "mm",
		true:  "in",
	}
	if cond.PrecipMM != nil {
		a := *cond.PrecipMM
		if c.imperial {
			a *= 0.039
		}

		if cond.ChanceOfRainPercent != nil {
			return aatPad(fmt.Sprintf("%.1f %s | %d%%", a, unit[c.imperial], *cond.ChanceOfRainPercent), 15)
		}
		return aatPad(fmt.Sprintf("%.1f %s", a, unit[c.imperial]), 15)
	} else if cond.ChanceOfRainPercent != nil {
		return aatPad(fmt.Sprintf("%d%%", *cond.ChanceOfRainPercent), 15)
	}
	return aatPad("", 15)
}

func (c *aatConfig) formatCond(cur []string, cond iface.Cond, current bool) (ret []string) {
	codes := map[iface.WeatherCode][]string{
		iface.CodeUnknown: {
			"    .-.      ",
			"     __)     ",
			"    (        ",
			"     `-’     ",
			"      •      ",
		},
		iface.CodeCloudy: {
			"             ",
			"\033[38;5;250m     .--.    \033[0m",
			"\033[38;5;250m  .-(    ).  \033[0m",
			"\033[38;5;250m (___.__)__) \033[0m",
			"             ",
		},
		iface.CodeFog: {
			"             ",
			"\033[38;5;251m _ - _ - _ - \033[0m",
			"\033[38;5;251m  _ - _ - _  \033[0m",
			"\033[38;5;251m _ - _ - _ - \033[0m",
			"             ",
		},
		iface.CodeHeavyRain: {
			"\033[38;5;240;1m     .-.     \033[0m",
			"\033[38;5;240;1m    (   ).   \033[0m",
			"\033[38;5;240;1m   (___(__)  \033[0m",
			"\033[38;5;21;1m  ‚‘‚‘‚‘‚‘   \033[0m",
			"\033[38;5;21;1m  ‚’‚’‚’‚’   \033[0m",
		},
		iface.CodeHeavyShowers: {
			"\033[38;5;226m _`/\"\"\033[38;5;240;1m.-.    \033[0m",
			"\033[38;5;226m  ,\\_\033[38;5;240;1m(   ).  \033[0m",
			"\033[38;5;226m   /\033[38;5;240;1m(___(__) \033[0m",
			"\033[38;5;21;1m   ‚‘‚‘‚‘‚‘  \033[0m",
			"\033[38;5;21;1m   ‚’‚’‚’‚’  \033[0m",
		},
		iface.CodeHeavySnow: {
			"\033[38;5;240;1m     .-.     \033[0m",
			"\033[38;5;240;1m    (   ).   \033[0m",
			"\033[38;5;240;1m   (___(__)  \033[0m",
			"\033[38;5;255;1m   * * * *   \033[0m",
			"\033[38;5;255;1m  * * * *    \033[0m",
		},
		iface.CodeHeavySnowShowers: {
			"\033[38;5;226m _`/\"\"\033[38;5;240;1m.-.    \033[0m",
			"\033[38;5;226m  ,\\_\033[38;5;240;1m(   ).  \033[0m",
			"\033[38;5;226m   /\033[38;5;240;1m(___(__) \033[0m",
			"\033[38;5;255;1m    * * * *  \033[0m",
			"\033[38;5;255;1m   * * * *   \033[0m",
		},
		iface.CodeLightRain: {
			"\033[38;5;250m     .-.     \033[0m",
			"\033[38;5;250m    (   ).   \033[0m",
			"\033[38;5;250m   (___(__)  \033[0m",
			"\033[38;5;111m    ‘ ‘ ‘ ‘  \033[0m",
			"\033[38;5;111m   ‘ ‘ ‘ ‘   \033[0m",
		},
		iface.CodeLightShowers: {
			"\033[38;5;226m _`/\"\"\033[38;5;250m.-.    \033[0m",
			"\033[38;5;226m  ,\\_\033[38;5;250m(   ).  \033[0m",
			"\033[38;5;226m   /\033[38;5;250m(___(__) \033[0m",
			"\033[38;5;111m     ‘ ‘ ‘ ‘ \033[0m",
			"\033[38;5;111m    ‘ ‘ ‘ ‘  \033[0m",
		},
		iface.CodeLightSleet: {
			"\033[38;5;250m     .-.     \033[0m",
			"\033[38;5;250m    (   ).   \033[0m",
			"\033[38;5;250m   (___(__)  \033[0m",
			"\033[38;5;111m    ‘ \033[38;5;255m*\033[38;5;111m ‘ \033[38;5;255m*  \033[0m",
			"\033[38;5;255m   *\033[38;5;111m ‘ \033[38;5;255m*\033[38;5;111m ‘   \033[0m",
		},
		iface.CodeLightSleetShowers: {
			"\033[38;5;226m _`/\"\"\033[38;5;250m.-.    \033[0m",
			"\033[38;5;226m  ,\\_\033[38;5;250m(   ).  \033[0m",
			"\033[38;5;226m   /\033[38;5;250m(___(__) \033[0m",
			"\033[38;5;111m     ‘ \033[38;5;255m*\033[38;5;111m ‘ \033[38;5;255m* \033[0m",
			"\033[38;5;255m    *\033[38;5;111m ‘ \033[38;5;255m*\033[38;5;111m ‘  \033[0m",
		},
		iface.CodeLightSnow: {
			"\033[38;5;250m     .-.     \033[0m",
			"\033[38;5;250m    (   ).   \033[0m",
			"\033[38;5;250m   (___(__)  \033[0m",
			"\033[38;5;255m    *  *  *  \033[0m",
			"\033[38;5;255m   *  *  *   \033[0m",
		},
		iface.CodeLightSnowShowers: {
			"\033[38;5;226m _`/\"\"\033[38;5;250m.-.    \033[0m",
			"\033[38;5;226m  ,\\_\033[38;5;250m(   ).  \033[0m",
			"\033[38;5;226m   /\033[38;5;250m(___(__) \033[0m",
			"\033[38;5;255m     *  *  * \033[0m",
			"\033[38;5;255m    *  *  *  \033[0m",
		},
		iface.CodePartlyCloudy: {
			"\033[38;5;226m   \\  /\033[0m      ",
			"\033[38;5;226m _ /\"\"\033[38;5;250m.-.    \033[0m",
			"\033[38;5;226m   \\_\033[38;5;250m(   ).  \033[0m",
			"\033[38;5;226m   /\033[38;5;250m(___(__) \033[0m",
			"             ",
		},
		iface.CodeSunny: {
			"\033[38;5;226m    \\   /    \033[0m",
			"\033[38;5;226m     .-.     \033[0m",
			"\033[38;5;226m  ― (   ) ―  \033[0m",
			"\033[38;5;226m     `-’     \033[0m",
			"\033[38;5;226m    /   \\    \033[0m",
		},
		iface.CodeThunderyHeavyRain: {
			"\033[38;5;240;1m     .-.     \033[0m",
			"\033[38;5;240;1m    (   ).   \033[0m",
			"\033[38;5;240;1m   (___(__)  \033[0m",
			"\033[38;5;21;1m  ‚‘\033[38;5;228;5m⚡\033[38;5;21;25m‘‚\033[38;5;228;5m⚡\033[38;5;21;25m‚‘   \033[0m",
			"\033[38;5;21;1m  ‚’‚’\033[38;5;228;5m⚡\033[38;5;21;25m’‚’   \033[0m",
		},
		iface.CodeThunderyShowers: {
			"\033[38;5;226m _`/\"\"\033[38;5;250m.-.    \033[0m",
			"\033[38;5;226m  ,\\_\033[38;5;250m(   ).  \033[0m",
			"\033[38;5;226m   /\033[38;5;250m(___(__) \033[0m",
			"\033[38;5;228;5m    ⚡\033[38;5;111;25m‘ ‘\033[38;5;228;5m⚡\033[38;5;111;25m‘ ‘ \033[0m",
			"\033[38;5;111m    ‘ ‘ ‘ ‘  \033[0m",
		},
		iface.CodeThunderySnowShowers: {
			"\033[38;5;226m _`/\"\"\033[38;5;250m.-.    \033[0m",
			"\033[38;5;226m  ,\\_\033[38;5;250m(   ).  \033[0m",
			"\033[38;5;226m   /\033[38;5;250m(___(__) \033[0m",
			"\033[38;5;255m     *\033[38;5;228;5m⚡\033[38;5;255;25m *\033[38;5;228;5m⚡\033[38;5;255;25m * \033[0m",
			"\033[38;5;255m    *  *  *  \033[0m",
		},
		iface.CodeVeryCloudy: {
			"             ",
			"\033[38;5;240;1m     .--.    \033[0m",
			"\033[38;5;240;1m  .-(    ).  \033[0m",
			"\033[38;5;240;1m (___.__)__) \033[0m",
			"             ",
		},
	}

	icon, ok := codes[cond.Code]
	if !ok {
		log.Fatalln("aat-frontend: The following weather code has no icon:", cond.Code)
	}

	desc := cond.Desc
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

func (c *aatConfig) printDay(day iface.Day) (ret []string) {
	desiredTimesOfDay := []time.Duration{
		8 * time.Hour,
		12 * time.Hour,
		19 * time.Hour,
		23 * time.Hour,
	}
	ret = make([]string, 5)
	for i := range ret {
		ret[i] = "│"
	}

	// save our selected elements from day.Slots in this array
	cols := make([]iface.Cond, len(desiredTimesOfDay))
	// find hourly data which fits the desired times of day best
	for _, candidate := range day.Slots {
		cand := candidate.Time.UTC().Sub(candidate.Time.Truncate(24 * time.Hour))
		for i, col := range cols {
			cur := col.Time.Sub(col.Time.Truncate(24 * time.Hour))
			if math.Abs(float64(cand-desiredTimesOfDay[i])) < math.Abs(float64(cur-desiredTimesOfDay[i])) {
				cols[i] = candidate
			}
		}
	}

	for _, s := range cols {
		ret = c.formatCond(ret, s, false)
		for i := range ret {
			ret[i] = ret[i] + "│"
		}
	}

	dateFmt := "┤ " + day.Date.Format("Mon 02. Jan") + " ├"
	ret = append([]string{
		"                                                       ┌─────────────┐                                                       ",
		"┌──────────────────────────────┬───────────────────────" + dateFmt + "───────────────────────┬──────────────────────────────┐",
		"│           Morning            │             Noon      └──────┬──────┘    Evening            │            Night             │",
		"├──────────────────────────────┼──────────────────────────────┼──────────────────────────────┼──────────────────────────────┤"},
		ret...)
	return append(ret,
		"└──────────────────────────────┴──────────────────────────────┴──────────────────────────────┴──────────────────────────────┘")
}

func (c *aatConfig) Setup() {
	flag.BoolVar(&c.imperial, "aat-imperial", false, "aat frontend: use imperial units for output")
}

func (c *aatConfig) Render(r iface.Data) {
	fmt.Printf("Weather for %s\n\n", r.Location)
	stdout := colorable.NewColorableStdout()

	out := c.formatCond(make([]string, 5), r.Current, true)
	for _, val := range out {
		fmt.Fprintln(stdout, val)
	}

	if len(r.Forecast) == 0 {
		return
	}
	if r.Forecast == nil {
		log.Fatal("No detailed weather forecast available.")
	}
	for _, d := range r.Forecast {
		for _, val := range c.printDay(d) {
			fmt.Fprintln(stdout, val)
		}
	}
}

func init() {
	iface.AllFrontends["ascii-art-table"] = &aatConfig{}
}
