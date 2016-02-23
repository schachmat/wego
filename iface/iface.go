package iface

import (
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

	// PrecipMM is the precipitation amount. It must be >= 0.
	PrecipMM *float32

	// VisibleDistKM is the visibility range in kilometers. It must be >= 0.
	VisibleDistKM *float32

	// WindspeedKmph is the average wind speed in kilometers per second.
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

	// MaxtempC is the maximum temperature on that day in degrees celsius.
	MaxtempC *float32

	// MintempC is the minimum temperature on that day in degrees celsius.
	MintempC *float32

	// Slots is a slice of conditions for different times of day. They should be
	// ordered by the contained Time field.
	Slots []Cond

	// Astronomy contains planetary data.
	Astronomy Astro
}

type Data struct {
	Current  Cond
	Forecast []Day
	Location string
}

type Backend interface {
	Setup()
	Fetch(location string, numdays int) Data
}

type Frontend interface {
	Setup()
	Render(weather Data)
}

var (
	AllBackends  = make(map[string]Backend)
	AllFrontends = make(map[string]Frontend)
)
