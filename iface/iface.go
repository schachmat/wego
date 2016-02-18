package iface

type Cond struct {
	ChanceOfRain   string  `json:"chanceofrain"`
	FeelsLikeC     int     `json:",string"`
	PrecipMM       float32 `json:"precipMM,string"`
	TempC          int     `json:"tempC,string"`
	TempC2         int     `json:"temp_C,string"`
	Time           int     `json:"time,string"`
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

type Weather struct {
	Astronomy []astro
	Date      string
	Hourly    []Cond
	MaxtempC  int `json:"maxtempC,string"`
	MintempC  int `json:"mintempC,string"`
}

type loc struct {
	Query string `json:"query"`
	Type  string `json:"type"`
}

type Resp struct {
	Data struct {
		Cur     []Cond                 `json:"current_condition"`
		Err     []struct{ Msg string } `json:"error"`
		Req     []loc                  `json:"request"`
		Weather []Weather              `json:"weather"`
	} `json:"data"`
}

type Backend interface {
	Setup()
	Fetch(location string, numdays int) Resp
}

type Frontend interface {
	Setup()
	Render(weather Resp)
}

var (
	AllBackends  = make(map[string]Backend)
	AllFrontends = make(map[string]Frontend)
)
