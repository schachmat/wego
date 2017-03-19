package iface

import (
	"log"
	"time"
)

type WeatherCode int

const (
	CodeUnknown WeatherCode = iota
	CodeCloudy
	CodeFog
	CodeHeavyRain
	CodeHeavyShowers
	CodeHeavySnow
	CodeHeavySnowShowers
	CodeLightRain
	CodeLightShowers
	CodeLightSleet
	CodeLightSleetShowers
	CodeLightSnow
	CodeLightSnowShowers
	CodePartlyCloudy
	CodeSunny
	CodeThunderyHeavyRain
	CodeThunderyShowers
	CodeThunderySnowShowers
	CodeVeryCloudy
)

type Cond struct {
	// Time is the time, where this weather condition applies.
	Time time.Time

	// Code is the general weather condition and must be one the WeatherCode
	// constants.
	Code WeatherCode

	// Desc is a short string describing the condition. It should be just one
	// sentence.
	Desc string

	// TempC is the temperature in degrees celsius.
	TempC *float32

	// FeelsLikeC is the felt temperature (with windchill effect e.g.) in
	// degrees celsius.
	FeelsLikeC *float32

	// ChanceOfRainPercent is the probability of rain or snow. It must be in the
	// range [0, 100].
	ChanceOfRainPercent *int

	// PrecipM is the precipitation amount in meters(!) per hour. Must be >= 0.
	PrecipM *float32

	// VisibleDistM is the visibility range in meters(!). It must be >= 0.
	VisibleDistM *float32

	// WindspeedKmph is the average wind speed in kilometers per hour. The value
	// must be >= 0.
	WindspeedKmph *float32

	// WindGustKmph is the maximum temporary wind speed in kilometers per
	// second. It should be > WindspeedKmph.
	WindGustKmph *float32

	// WinddirDegree is the direction the wind is blowing from on a clock
	// oriented circle with 360 degrees. 0 means the wind is blowing from north,
	// 90 means the wind is blowing from east, 180 means the wind is blowing
	// from south and 270 means the wind is blowing from west. The value must be
	// in the range [0, 359].
	WinddirDegree *int

	// Humidity is the *relative* humidity and must be in [0, 100].
	Humidity *int
}

type Astro struct {
	Moonrise time.Time
	Moonset  time.Time
	Sunrise  time.Time
	Sunset   time.Time
}

type Day struct {
	// Date is the date of this Day.
	Date time.Time

	// Slots is a slice of conditions for different times of day. They should be
	// ordered by the contained Time field.
	Slots []Cond

	// Astronomy contains planetary data.
	Astronomy Astro
}

type LatLon struct {
	Latitude  float32
	Longitude float32
}

type Data struct {
	Current  Cond
	Forecast []Day
	Location string
	GeoLoc   *LatLon
}

type UnitSystem int

const (
	UnitsMetric UnitSystem = iota
	UnitsImperial
	UnitsSi
	UnitsMetricMs
)

func (u UnitSystem) Temp(tempC float32) (res float32, unit string) {
	if u == UnitsMetric || u == UnitsMetricMs {
		return tempC, "°C"
	} else if u == UnitsImperial {
		return tempC*1.8 + 32, "°F"
	} else if u == UnitsSi {
		return tempC + 273.16, "°K"
	}
	log.Fatalln("Unknown unit system:", u)
	return
}

func (u UnitSystem) Speed(spdKmph float32) (res float32, unit string) {
	if u == UnitsMetric {
		return spdKmph, "km/h"
	} else if u == UnitsImperial {
		return spdKmph / 1.609, "mph"
	} else if u == UnitsSi || u == UnitsMetricMs {
		return spdKmph / 3.6, "m/s"
	}
	log.Fatalln("Unknown unit system:", u)
	return
}

func (u UnitSystem) Distance(distM float32) (res float32, unit string) {
	if u == UnitsMetric || u == UnitsSi || u == UnitsMetricMs {
		if distM < 1 {
			return distM * 1000, "mm"
		} else if distM < 1000 {
			return distM, "m"
		} else {
			return distM / 1000, "km"
		}
	} else if u == UnitsImperial {
		res, unit = distM/0.0254, "in"
		if res < 3*12 { // 1yd = 3ft, 1ft = 12in
			return
		} else if res < 8*10*22*36 { //1mi = 8fur, 1fur = 10ch, 1ch = 22yd
			return res / 36, "yd"
		} else {
			return res / 8 / 10 / 22 / 36, "mi"
		}
	}
	log.Fatalln("Unknown unit system:", u)
	return
}

type Backend interface {
	Setup()
	Fetch(location string, numdays int) Data
}

type Frontend interface {
	Setup()
	Render(weather Data, unitSystem UnitSystem)
}

var (
	AllBackends  = make(map[string]Backend)
	AllFrontends = make(map[string]Frontend)
)
