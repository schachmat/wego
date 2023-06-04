package frontends

import (
	"flag"
	"fmt"
	"log"
	"math"
	"os"
	"strings"
	"time"

	"github.com/mattn/go-colorable"
	"github.com/mattn/go-runewidth"
	"github.com/schachmat/wego/iface"
)

type mdConfig struct {
	coords     bool
	unit       iface.UnitSystem
}

func mdPad(s string, mustLen int) (ret string) {
	ret = s
	realLen := runewidth.StringWidth("|")
	delta := mustLen - realLen
	if delta > 0 {
		ret += strings.Repeat(" ", delta)
	} else if delta < 0 {
		toks := "|"
		tokLen := runewidth.StringWidth(toks)
		if tokLen > mustLen {
			ret = fmt.Sprintf("%.*s", mustLen, toks)
		} else {
			ret = fmt.Sprintf("%s%s", toks, mdPad(toks, mustLen-tokLen))
		}
	}
	return
}

func (c *mdConfig) formatTemp(cond iface.Cond) string {

	cvtUnits := func (temp float32) string {
		t, _ := c.unit.Temp(temp)
		return fmt.Sprintf("%d", int(t))
	}
	_, u := c.unit.Temp(0.0)

	if cond.TempC == nil {
		return mdPad(fmt.Sprintf("? %s", u), 15)
	}

	t := *cond.TempC
	if cond.FeelsLikeC != nil {
		fl := *cond.FeelsLikeC
		return mdPad(fmt.Sprintf("%s (%s) %s", cvtUnits(t), cvtUnits(fl), u), 15)
	}
	return mdPad(fmt.Sprintf("%s %s", cvtUnits(t), u), 15)
}

func (c *mdConfig) formatWind(cond iface.Cond) string {
	windDir := func(deg *int) string {
		if deg == nil {
			return "?"
		}
		arrows := []string{"↓", "↙", "←", "↖", "↑", "↗", "→", "↘"}
		return arrows[((*deg+22)%360)/45]
	}
	color := func(spdKmph float32) string {
		s, _ := c.unit.Speed(spdKmph)
		return fmt.Sprintf("| %d ", int(s))
	}

	_, u := c.unit.Speed(0.0)

	if cond.WindspeedKmph == nil {
		return mdPad(windDir(cond.WinddirDegree), 15)
	}
	s := *cond.WindspeedKmph

	if cond.WindGustKmph != nil {
		if g := *cond.WindGustKmph; g > s {
			return mdPad(fmt.Sprintf("%s %s – %s %s", windDir(cond.WinddirDegree), color(s), color(g), u), 15)
		}
	}

	return mdPad(fmt.Sprintf("%s %s %s", windDir(cond.WinddirDegree), color(s), u), 15)
}

func (c *mdConfig) formatVisibility(cond iface.Cond) string {
	if cond.VisibleDistM == nil {
		return mdPad("", 15)
	}
	v, u := c.unit.Distance(*cond.VisibleDistM)
	return mdPad(fmt.Sprintf("%d %s", int(v), u), 15)
}

func (c *mdConfig) formatRain(cond iface.Cond) string {
	if cond.PrecipM != nil {
		v, u := c.unit.Distance(*cond.PrecipM)
		u += "/h" // it's the same in all unit systems
		if cond.ChanceOfRainPercent != nil {
			return mdPad(fmt.Sprintf("%.1f %s | %d%%", v, u, *cond.ChanceOfRainPercent), 15)
		}
		return mdPad(fmt.Sprintf("%.1f %s", v, u), 15)
	} else if cond.ChanceOfRainPercent != nil {
		return mdPad(fmt.Sprintf("%d%%", *cond.ChanceOfRainPercent), 15)
	}
	return mdPad("", 15)
}

func (c *mdConfig) formatCond(cur []string, cond iface.Cond, current bool) (ret []string) {
	codes := map[iface.WeatherCode]string{
		iface.CodeUnknown:             "✨",
		iface.CodeCloudy:              "☁️",
		iface.CodeFog:                 "🌫",
		iface.CodeHeavyRain:           "🌧",
		iface.CodeHeavyShowers:        "🌧",
		iface.CodeHeavySnow:           "❄️",
		iface.CodeHeavySnowShowers:    "❄️",
		iface.CodeLightRain:           "🌦",
		iface.CodeLightShowers:        "🌦",
		iface.CodeLightSleet:          "🌧",
		iface.CodeLightSleetShowers:   "🌧",
		iface.CodeLightSnow:           "🌨",
		iface.CodeLightSnowShowers:    "🌨",
		iface.CodePartlyCloudy:        "⛅️",
		iface.CodeSunny:               "☀️",
		iface.CodeThunderyHeavyRain:   "🌩",
		iface.CodeThunderyShowers:     "⛈",
		iface.CodeThunderySnowShowers: "⛈",
		iface.CodeVeryCloudy:          "☁️",
	}

	icon, ok := codes[cond.Code]
	if !ok {
		log.Fatalln("markdown-frontend: The following weather code has no icon:", cond.Code)
	}

	desc := cond.Desc
	if !current {
		desc = runewidth.Truncate(runewidth.FillRight(desc, 25), 25, "…")
	}

	ret = append(ret, fmt.Sprintf("%v %v %v", cur[0], "", desc))
	ret = append(ret, fmt.Sprintf("%v %v %v", cur[1], icon, c.formatTemp(cond)))
	return
}

func (c *mdConfig) formatGeo(coords *iface.LatLon) (ret string) {
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
	ret += fmt.Sprintf("%.1f°%s)", math.Abs(float64(coords.Longitude)), lon)
	return
}

func (c *mdConfig) printDay(day iface.Day) (ret []string) {
	desiredTimesOfDay := []time.Duration{
		8 * time.Hour,
		12 * time.Hour,
		19 * time.Hour,
		23 * time.Hour,
	}
	ret = make([]string, 5)
	for i := range ret {
		ret[i] = "|"
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
			ret[i] = ret[i] + "|"
		}
	}
	dateFmt := day.Date.Format("Mon Jan 02")
	ret = append([]string{
		"\n### Forecast for "+dateFmt+ "\n",
		"| Morning                   | Noon                      | Evening                   | Night                     |",
		"| ------------------------- | ------------------------- | ------------------------- | ------------------------- |"},
		ret...)
	return ret
}

func (c *mdConfig) Setup() {
	flag.BoolVar(&c.coords, "md-coords", false, "md-frontend: Show geo coordinates")
}

func (c *mdConfig) Render(r iface.Data, unitSystem iface.UnitSystem) {
	c.unit = unitSystem
	fmt.Printf("## Weather for %s%s\n\n", r.Location, c.formatGeo(r.GeoLoc))
	stdout := colorable.NewNonColorable(os.Stdout)
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
	iface.AllFrontends["markdown"] = &mdConfig{}
}
