package frontends

import (
	"fmt"
	"log"
	"math"
	"time"

	colorable "github.com/mattn/go-colorable"
	runewidth "github.com/mattn/go-runewidth"
	"github.com/schachmat/wego/iface"
)

type emojiConfig struct {
	unit iface.UnitSystem
}

func (c *emojiConfig) formatTemp(cond iface.Cond) string {
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
		return aatPad(fmt.Sprintf("? %s", u), 12)
	}

	t := *cond.TempC
	if cond.FeelsLikeC != nil {
		fl := *cond.FeelsLikeC
		return aatPad(fmt.Sprintf("%s (%s) %s", color(t), color(fl), u), 12)
	}
	return aatPad(fmt.Sprintf("%s %s", color(t), u), 12)
}

func (c *emojiConfig) formatCond(cur []string, cond iface.Cond, current bool) (ret []string) {
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
		log.Fatalln("emoji-frontend: The following weather code has no icon:", cond.Code)
	}
	if runewidth.StringWidth(icon) == 1 {
		icon += " "
	}

	desc := cond.Desc
	if !current {
		desc = runewidth.Truncate(runewidth.FillRight(desc, 13), 13, "…")
	}

	ret = append(ret, fmt.Sprintf("%v %v %v", cur[0], "", desc))
	ret = append(ret, fmt.Sprintf("%v%v %v", cur[1], icon, c.formatTemp(cond)))
	return
}

func (c *emojiConfig) printAstro(astro iface.Astro) {
        // print sun astronomy data if present
	if astro.Sunrise != astro.Sunset {
	    // half the distance between sunrise and sunset
	    noon_distance := time.Duration(int64(float32(astro.Sunset.UnixNano() - astro.Sunrise.UnixNano()) * 0.5))
	    // time for solar noon
	    noon := astro.Sunrise.Add(noon_distance)

	    // the actual print statement
	    fmt.Printf("🌞 rise↗ %s noon↑ %s set↘ %s\n", astro.Sunrise.Format(time.Kitchen), noon.Format(time.Kitchen), astro.Sunset.Format(time.Kitchen))
	}
        // print moon astronomy data if present
	if astro.Moonrise != astro.Moonset {
	    fmt.Printf("🌚 rise↗ %s set↘ %s\n", astro.Moonrise.Format(time.Kitchen), astro.Moonset)
	}
}

func (c *emojiConfig) printDay(day iface.Day) (ret []string) {
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

	c.printAstro(day.Astronomy)

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

	dateFmt := "┤  " + day.Date.Format("Mon") + "  ├"
	ret = append([]string{
		"                            ┌───────┐ ",
		"┌───────────────┬───────────" + dateFmt + "───────────┬───────────────┐",
		"│    Morning    │    Noon   └───┬───┘ Evening   │     Night     │",
		"├───────────────┼───────────────┼───────────────┼───────────────┤"},
		ret...)
	return append(ret,
		"└───────────────┴───────────────┴───────────────┴───────────────┘",
		" ")
}

func (c *emojiConfig) Setup() {
}

func (c *emojiConfig) Render(r iface.Data, unitSystem iface.UnitSystem) {
	c.unit = unitSystem

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
	fmt.Printf("\n")
	for _, d := range r.Forecast {
		for _, val := range c.printDay(d) {
			fmt.Fprintln(stdout, val)
		}
	}
}

func init() {
	iface.AllFrontends["emoji"] = &emojiConfig{}
}
