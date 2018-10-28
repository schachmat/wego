package frontends

import (
	"flag"
	"fmt"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/gomono"
	"io/ioutil"
	"log"
	"math"
	"time"

	"github.com/fogleman/gg"
	"github.com/mattn/go-runewidth"
	"github.com/schachmat/wego/iface"
)

type pngimageConfig struct {
	unit iface.UnitSystem
	fontPath string
}

func truncateString(str string, num int) string {
	bnoden := str
	if len(str) > num {
		if num > 1 {
			num -= 1
		}
		bnoden = str[0:num] + "…"
	}
	return bnoden
}

func (c *pngimageConfig) formatCond(dc *gg.Context, cond iface.Cond, current bool, baseX float64, baseY float64) {
	lineHeight := float64(20)
	firstLine := float64(-15)
	codes := map[iface.WeatherCode]uint32{
		iface.CodeUnknown:             '\uf07b',
		iface.CodeCloudy:              '\uf002',
		iface.CodeFog:                 '\uf003',
		iface.CodeHeavyRain:           '\uf008',
		iface.CodeHeavyShowers:        '\uf009',
		iface.CodeHeavySnow:           '\uf00a',
		iface.CodeHeavySnowShowers:    '\uf065',
		iface.CodeLightRain:           '\uf006',
		iface.CodeLightShowers:        '\uf009',
		iface.CodeLightSleet:          '\uf0b2',
		iface.CodeLightSleetShowers:   '\uf0b2',
		iface.CodeLightSnow:           '\uf065',
		iface.CodeLightSnowShowers:    '\uf00a',
		iface.CodePartlyCloudy:        '\uf002',
		iface.CodeSunny:               '\uf00d',
		iface.CodeThunderyHeavyRain:   '\uf010',
		iface.CodeThunderyShowers:     '\uf00e',
		iface.CodeThunderySnowShowers: '\uf06b',
		iface.CodeVeryCloudy:          '\uf013',
	}

	icon, ok := codes[cond.Code]
	if !ok {
		log.Fatalln("pngimage-frontend: The following weather code has no icon:", cond.Code)
	}

	currentHour := cond.Time.Hour()
	if cond.Code != iface.CodeUnknown && (currentHour < 5 || currentHour > 22) {
		icon = icon + 0x21
	}

	desc := cond.Desc
	if !current {
		desc = runewidth.Truncate(runewidth.FillRight(desc, 13), 13, "…")
	}

	dc.SetRGB(0, 0, 0)
	dc.SetFontFace(weatherFont)
	dc.DrawString(fmt.Sprintf("%c", icon), baseX, baseY)
	dc.SetFontFace(smallTextFont)
	dc.DrawString(truncateString(cond.Desc, 15), baseX+50, baseY+firstLine)

	if cond.TempC != nil {
		_, u := c.unit.Temp(0.0)

		dc.DrawString(fmt.Sprintf("%.1f %s", *cond.TempC, u), baseX+50, baseY+firstLine+lineHeight*1)
	}

	if cond.WindspeedKmph != nil {
		_, u := c.unit.Speed(0.0)

		direction := "\uf07b"
		if cond.WinddirDegree != nil {
			arrows := []string{"\uf044", "\uf043", "\uf048", "\uf087", "\uf058", "\uf057", "\uf04d", "\uf088"}
			direction = arrows[((*cond.WinddirDegree+22)%360)/45]
		}

		dc.SetFontFace(weatherFont)
		dc.DrawString(fmt.Sprintf("%s", direction), baseX+35, baseY+firstLine+lineHeight*2+5)
		dc.SetFontFace(smallTextFont)
		dc.DrawString(fmt.Sprintf("%.1f %s", *cond.WindspeedKmph, u), baseX+50, baseY+firstLine+lineHeight*2)
	}

	if cond.VisibleDistM != nil {
		d, u := c.unit.Distance(*cond.VisibleDistM)

		dc.DrawString(fmt.Sprintf("%.1f %s", d, u), baseX+50, baseY+firstLine+lineHeight*3)
	}

	if cond.PrecipM != nil {
		v, u := c.unit.Distance(*cond.PrecipM)
		u += "/h" // it's the same in all unit systems
		if cond.ChanceOfRainPercent != nil {
			dc.DrawString(fmt.Sprintf("%.1f %s | %d%%", v, u, *cond.ChanceOfRainPercent), baseX+50, baseY+firstLine+lineHeight*4)
		} else {
			dc.DrawString(fmt.Sprintf("%.1f %s", v, u), baseX+50, baseY+firstLine+lineHeight*4)
		}
	} else if cond.ChanceOfRainPercent != nil {
		dc.DrawString(fmt.Sprintf("%d%%", *cond.ChanceOfRainPercent), baseX+50, baseY+firstLine+lineHeight*4)
	}

	return
}

func (c *pngimageConfig) printDay(dc *gg.Context, day iface.Day, baseX float64, baseY float64) {
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

	dc.SetLineWidth(2)
	dc.DrawLine(baseX, baseY+10, baseX+float64(200*len(cols)), baseY+10)
	dc.Stroke()

	dc.SetFontFace(textFont)
	dc.DrawString(day.Date.Format("Mon 02 Jan"), baseX, baseY)

	for j, s := range cols {
		c.formatCond(dc, s, false, baseX+float64(200*j), baseY+50)
	}
}

func loadFontFace(path string, points float64) (font.Face, error) {
	fontBytes, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	f, err := truetype.Parse(fontBytes)
	if err != nil {
		return nil, err
	}
	face := truetype.NewFace(f, &truetype.Options{
		Size: points,
		// Hinting: font.HintingFull,
	})
	return face, nil
}

func loadGoFontFace(size float64) (font.Face, error) {

	font, err := truetype.Parse(gomono.TTF)
	if err != nil {
		return nil, err
	}

	face := truetype.NewFace(font, &truetype.Options{Size: size})
	return face, nil
}

var (
	weatherFont   font.Face
	textFont      font.Face
	smallTextFont font.Face

	desiredTimesOfDay = []time.Duration{
		8 * time.Hour,
		12 * time.Hour,
		19 * time.Hour,
		23 * time.Hour,
	}
)

func (c *pngimageConfig) Setup() {
	flag.StringVar(&c.fontPath, "pngimage-font", "weathericons-regular-webfont.ttf", "pngimage frontend: the path to the weather font")
}

func (c *pngimageConfig) Render(r iface.Data, unitSystem iface.UnitSystem) {
	c.unit = unitSystem

	if c.fontPath == "" {
		log.Fatalf("No font specified. Please download the ttf font at http://weathericons.io and link to it in the configuration file.")
	}

	dc := gg.NewContext(200*len(desiredTimesOfDay)+200, 150*len(r.Forecast)+250)
	dc.SetRGBA(1, 1, 1, 0.5)
	dc.Clear()

	var err error
	weatherFont, err = loadFontFace(c.fontPath, 24)
	if err != nil {
		log.Fatalf("Invalid font specified. Please download the ttf font at http://weathericons.io and link to it in the configuration file.")
	}

	textFont, _ = loadGoFontFace(20)
	smallTextFont, _ = loadGoFontFace(14)

	dc.SetHexColor("#00000")
	dc.SetFontFace(textFont)
	dc.DrawString(fmt.Sprintf("Weather for %s", r.Location), 100, 30)

	c.formatCond(dc, r.Current, true, 100, 100)

	if len(r.Forecast) == 0 {
		return
	}
	if r.Forecast == nil {
		log.Fatal("No detailed weather forecast available.")
	}
	for i, d := range r.Forecast {
		c.printDay(dc, d, 100, 225+float64(i*150))
	}
	dc.SavePNG("out.png")
}

func init() {
	iface.AllFrontends["pngimage"] = &pngimageConfig{}
}
