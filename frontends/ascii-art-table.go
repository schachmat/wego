package frontends

import (
	"flag"
	"fmt"
	"log"
	"math"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/mattn/go-colorable"
	"github.com/mattn/go-runewidth"
	"github.com/schachmat/wego/iface"
)

type aatConfig struct {
	coords     bool
	monochrome bool
	compact    bool

	unit iface.UnitSystem
}

// TODO: replace s parameter with printf interface?
func aatPad(s string, mustLen int) (ret string) {
	ansiEsc := regexp.MustCompile("\033.*?m")
	ret = s
	realLen := runewidth.StringWidth(ansiEsc.ReplaceAllLiteralString(s, ""))
	delta := mustLen - realLen
	if delta > 0 {
		ret += "\033[0m" + strings.Repeat(" ", delta)
	} else if delta < 0 {
		toks := ansiEsc.Split(s, 2)
		tokLen := runewidth.StringWidth(toks[0])
		if tokLen > mustLen {
			ret = fmt.Sprintf("%.*s\033[0m", mustLen, toks[0])
		} else {
			esc := ansiEsc.FindString(s)
			ret = fmt.Sprintf("%s%s%s", toks[0], esc, aatPad(toks[1], mustLen-tokLen))
		}
	}
	return
}

func (c *aatConfig) formatTemp(cond iface.Cond) string {
	color := func(temp float32) string {
		colmap := []struct {
			maxtemp float32
			color   int
		}{
			{-15, 21}, {-12, 27}, {-9, 33}, {-6, 39}, {-3, 45},
			{0, 51}, {2, 50}, {4, 49}, {6, 48}, {8, 47},
			{10, 46}, {13, 82}, {16, 118}, {19, 154}, {22, 190},
			{25, 226}, {28, 220}, {31, 214}, {34, 208}, {37, 202},
		}

		col := 196
		for _, candidate := range colmap {
			if temp < candidate.maxtemp {
				col = candidate.color
				break
			}
		}
		t, _ := c.unit.Temp(temp)
		return fmt.Sprintf("\033[38;5;%03dm%d\033[0m", col, int(t))
	}

	_, u := c.unit.Temp(0.0)

	if cond.TempC == nil {
		return aatPad(fmt.Sprintf("? %s", u), 15)
	}

	t := *cond.TempC
	if cond.FeelsLikeC != nil {
		fl := *cond.FeelsLikeC
		return aatPad(fmt.Sprintf("%s (%s) %s", color(t), color(fl), u), 15)
	}
	return aatPad(fmt.Sprintf("%s %s", color(t), u), 15)
}

func (c *aatConfig) formatWind(cond iface.Cond) string {
	windDir := func(deg *int) string {
		if deg == nil {
			return "?"
		}
		arrows := []string{"↓", "↙", "←", "↖", "↑", "↗", "→", "↘"}
		return "\033[1m" + arrows[((*deg+22)%360)/45] + "\033[0m"
	}
	color := func(spdKmph float32) string {
		colmap := []struct {
			maxtemp float32
			color   int
		}{
			{0, 46}, {4, 82}, {7, 118}, {10, 154}, {13, 190},
			{16, 226}, {20, 220}, {24, 214}, {28, 208}, {32, 202},
		}

		col := 196
		for _, candidate := range colmap {
			if spdKmph < candidate.maxtemp {
				col = candidate.color
				break
			}
		}

		s, _ := c.unit.Speed(spdKmph)
		return fmt.Sprintf("\033[38;5;%03dm%d\033[0m", col, int(s))
	}

	_, u := c.unit.Speed(0.0)

	if cond.WindspeedKmph == nil {
		return aatPad(windDir(cond.WinddirDegree), 15)
	}
	s := *cond.WindspeedKmph

	if cond.WindGustKmph != nil {
		if g := *cond.WindGustKmph; g > s {
			return aatPad(fmt.Sprintf("%s %s – %s %s", windDir(cond.WinddirDegree), color(s), color(g), u), 15)
		}
	}

	return aatPad(fmt.Sprintf("%s %s %s", windDir(cond.WinddirDegree), color(s), u), 15)
}

func (c *aatConfig) formatVisibility(cond iface.Cond) string {
	if cond.VisibleDistM == nil {
		return aatPad("", 15)
	}
	v, u := c.unit.Distance(*cond.VisibleDistM)
	return aatPad(fmt.Sprintf("%d %s", int(v), u), 15)
}

func (c *aatConfig) formatRain(cond iface.Cond) string {
	if cond.PrecipM != nil {
		v, u := c.unit.Distance(*cond.PrecipM)
		u += "/h" // it's the same in all unit systems
		if cond.ChanceOfRainPercent != nil {
			return aatPad(fmt.Sprintf("%.1f %s | %d%%", v, u, *cond.ChanceOfRainPercent), 15)
		}
		return aatPad(fmt.Sprintf("%.1f %s", v, u), 15)
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
			"     `-᾿     ",
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
			"\033[38;5;244;1m     .-.     \033[0m",
			"\033[38;5;244;1m    (   ).   \033[0m",
			"\033[38;5;244;1m   (___(__)  \033[0m",
			"\033[38;5;33;1m  ‚ʻ‚ʻ‚ʻ‚ʻ   \033[0m",
			"\033[38;5;33;1m  ‚ʻ‚ʻ‚ʻ‚ʻ   \033[0m",
		},
		iface.CodeHeavyShowers: {
			"\033[38;5;226m _`/\"\"\033[38;5;244;1m.-.    \033[0m",
			"\033[38;5;226m  ,\\_\033[38;5;244;1m(   ).  \033[0m",
			"\033[38;5;226m   /\033[38;5;244;1m(___(__) \033[0m",
			"\033[38;5;33;1m   ‚ʻ‚ʻ‚ʻ‚ʻ  \033[0m",
			"\033[38;5;33;1m   ‚ʻ‚ʻ‚ʻ‚ʻ  \033[0m",
		},
		iface.CodeHeavySnow: {
			"\033[38;5;244;1m     .-.     \033[0m",
			"\033[38;5;244;1m    (   ).   \033[0m",
			"\033[38;5;244;1m   (___(__)  \033[0m",
			"\033[38;5;255;1m   * * * *   \033[0m",
			"\033[38;5;255;1m  * * * *    \033[0m",
		},
		iface.CodeHeavySnowShowers: {
			"\033[38;5;226m _`/\"\"\033[38;5;244;1m.-.    \033[0m",
			"\033[38;5;226m  ,\\_\033[38;5;244;1m(   ).  \033[0m",
			"\033[38;5;226m   /\033[38;5;244;1m(___(__) \033[0m",
			"\033[38;5;255;1m    * * * *  \033[0m",
			"\033[38;5;255;1m   * * * *   \033[0m",
		},
		iface.CodeLightRain: {
			"\033[38;5;250m     .-.     \033[0m",
			"\033[38;5;250m    (   ).   \033[0m",
			"\033[38;5;250m   (___(__)  \033[0m",
			"\033[38;5;111m    ʻ ʻ ʻ ʻ  \033[0m",
			"\033[38;5;111m   ʻ ʻ ʻ ʻ   \033[0m",
		},
		iface.CodeLightShowers: {
			"\033[38;5;226m _`/\"\"\033[38;5;250m.-.    \033[0m",
			"\033[38;5;226m  ,\\_\033[38;5;250m(   ).  \033[0m",
			"\033[38;5;226m   /\033[38;5;250m(___(__) \033[0m",
			"\033[38;5;111m     ʻ ʻ ʻ ʻ \033[0m",
			"\033[38;5;111m    ʻ ʻ ʻ ʻ  \033[0m",
		},
		iface.CodeLightSleet: {
			"\033[38;5;250m     .-.     \033[0m",
			"\033[38;5;250m    (   ).   \033[0m",
			"\033[38;5;250m   (___(__)  \033[0m",
			"\033[38;5;111m    ʻ \033[38;5;255m*\033[38;5;111m ʻ \033[38;5;255m*  \033[0m",
			"\033[38;5;255m   *\033[38;5;111m ʻ \033[38;5;255m*\033[38;5;111m ʻ   \033[0m",
		},
		iface.CodeLightSleetShowers: {
			"\033[38;5;226m _`/\"\"\033[38;5;250m.-.    \033[0m",
			"\033[38;5;226m  ,\\_\033[38;5;250m(   ).  \033[0m",
			"\033[38;5;226m   /\033[38;5;250m(___(__) \033[0m",
			"\033[38;5;111m     ʻ \033[38;5;255m*\033[38;5;111m ʻ \033[38;5;255m* \033[0m",
			"\033[38;5;255m    *\033[38;5;111m ʻ \033[38;5;255m*\033[38;5;111m ʻ  \033[0m",
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
			"\033[38;5;226m   \\__/\033[0m      ",
			"\033[38;5;226m __/  \033[38;5;250m.-.    \033[0m",
			"\033[38;5;226m   \\_\033[38;5;250m(   ).  \033[0m",
			"\033[38;5;226m   /\033[38;5;250m(___(__) \033[0m",
			"             ",
		},
		iface.CodeSunny: {
			"\033[38;5;226m    \\ . /    \033[0m",
			"\033[38;5;226m   - .-. -   \033[0m",
			"\033[38;5;226m  ‒ (   ) ‒  \033[0m",
			"\033[38;5;226m   . `-᾿ .   \033[0m",
			"\033[38;5;226m    / ' \\    \033[0m",
		},
		iface.CodeThunderyHeavyRain: {
			"\033[38;5;244;1m     .-.     \033[0m",
			"\033[38;5;244;1m    (   ).   \033[0m",
			"\033[38;5;244;1m   (___(__)  \033[0m",
			"\033[38;5;33;1m  ‚ʻ\033[38;5;228;5m⚡\033[38;5;33;25mʻ‚\033[38;5;228;5m⚡\033[38;5;33;25m‚ʻ   \033[0m",
			"\033[38;5;33;1m  ‚ʻ‚ʻ\033[38;5;228;5m⚡\033[38;5;33;25mʻ‚ʻ   \033[0m",
		},
		iface.CodeThunderyShowers: {
			"\033[38;5;226m _`/\"\"\033[38;5;250m.-.    \033[0m",
			"\033[38;5;226m  ,\\_\033[38;5;250m(   ).  \033[0m",
			"\033[38;5;226m   /\033[38;5;250m(___(__) \033[0m",
			"\033[38;5;228;5m    ⚡\033[38;5;111;25mʻ ʻ\033[38;5;228;5m⚡\033[38;5;111;25mʻ ʻ \033[0m",
			"\033[38;5;111m    ʻ ʻ ʻ ʻ  \033[0m",
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
			"\033[38;5;244;1m     .--.    \033[0m",
			"\033[38;5;244;1m  .-(    ).  \033[0m",
			"\033[38;5;244;1m (___.__)__) \033[0m",
			"             ",
		},
	}

	icon := make([]string, 5)
	if !c.compact {
		var ok bool
		icon, ok = codes[cond.Code]
		if !ok {
			log.Fatalln("aat-frontend: The following weather code has no icon:", cond.Code)
		}
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

func (c *aatConfig) formatGeo(coords *iface.LatLon) (ret string) {
	if !c.coords || coords == nil {
		return ""
	}

	lat, lon := "N", "E"
	if coords.Latitude < 0 {
		lat = "S"
	}
	if coords.Longitude < 0 {
		lon = "W"
	}
	ret = " "
	ret += fmt.Sprintf("(%.1f°%s", math.Abs(float64(coords.Latitude)), lat)
	ret += fmt.Sprintf(" %.1f°%s)", math.Abs(float64(coords.Longitude)), lon)
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
			if col.Time.IsZero() || math.Abs(float64(cand-desiredTimesOfDay[i])) < math.Abs(float64(cur-desiredTimesOfDay[i])) {
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
	if !c.compact {
		ret = append([]string{
			"                                                       ┌─────────────┐                                                       ",
			"┌──────────────────────────────┬───────────────────────" + dateFmt + "───────────────────────┬──────────────────────────────┐",
			"│           Morning            │             Noon      └──────┬──────┘    Evening            │            Night             │",
			"├──────────────────────────────┼──────────────────────────────┼──────────────────────────────┼──────────────────────────────┤"},
			ret...)
		ret = append(ret,
			"└──────────────────────────────┴──────────────────────────────┴──────────────────────────────┴──────────────────────────────┘")
	} else {
		merge := func(src string, into string) string {
			ret := []rune(into)
			for k, v := range src {
				ret[k] = v
			}
			return string(ret)
		}

		spaces := (len(ret[0]) / 4) - 3
		bar := strings.Repeat("─", spaces)

		ret = append([]string{
			day.Date.Format("Mon 02. Jan"),
			"┌" + merge("Morning", bar) + "┬" + merge("Noon", bar) + "┬" + merge("Evening", bar) + "┬" + merge("Night", bar) + "┐",
		}, ret...)

		ret = append(ret,
			"└"+bar+"┴"+bar+"┴"+bar+"┴"+bar+"┘",
		)
	}

	return ret
}

func (c *aatConfig) Setup() {
	flag.BoolVar(&c.coords, "aat-coords", false, "aat-frontend: Show geo coordinates")
	flag.BoolVar(&c.monochrome, "aat-monochrome", false, "aat-frontend: Monochrome output")

	flag.BoolVar(&c.compact, "aat-compact", false, "aat-frontend: Compact output")
}

func (c *aatConfig) Render(r iface.Data, unitSystem iface.UnitSystem) {
	c.unit = unitSystem

	fmt.Printf("Weather for %s%s\n\n", r.Location, c.formatGeo(r.GeoLoc))
	stdout := colorable.NewColorableStdout()
	if c.monochrome {
		stdout = colorable.NewNonColorable(os.Stdout)
	}

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
