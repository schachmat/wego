package main

type configuration struct {
	APIKey   string
	City     string
	Imperial bool
}

type cond struct {
	ChanceOfRain   string  `json:"chanceofrain"`
	FeelsLikeC     int     `json:",string"`
	PrecipMM       float32 `json:"precipMM,string"`
	TempC          int     `json:"tempC,string"`
	Time           string  `json:"time"`
	VisibleDistKM  int     `json:"visibility,string"`
	WeatherCode    int     `json:"weatherCode,string"`
	WeatherDesc    []struct{ Value string }
	WindGustKmph   int `json:",string"`
	Winddir16Point string
	WindspeedKmph  int `json:"windspeedKmph,string"`
}

type astro struct {
	Moonrise string
	Moonset  string
	Sunrise  string
	Sunset   string
}

type weather struct {
	Astronomy []astro
	Date      string
	Hourly    []cond
	MaxtempC  int `json:"maxtempC,string"`
	MintempC  int `json:"mintempC,string"`
}

type loc struct {
	Query string `json:"query"`
	Type  string `json:"type"`
}

type resp struct {
	Data struct {
		Cur     []cond                 `json:"current_condition"`
		Err     []struct{ Msg string } `json:"error"`
		Req     []loc                  `json:"request"`
		Weather []weather              `json:"weather"`
	} `json:"data"`
}

var (
	config     configuration
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
		395: iconHeavySnowShowers, // ThunderyHeavySnow
	}
)
