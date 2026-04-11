package backends

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/schachmat/wego/iface"
)

type weatherXuConfig struct {
	apiKey string
	lang   string
	debug  bool
}

// WeatherXuResponse represents the main response structure
type weatherXuResponse struct {
	Success bool      `json:"success"`
	Error   ErrorData `json:"error,omitempty"`
	Data    struct {
		Dt                   int64   `json:"dt"`
		Latitude             float64 `json:"latitude"`
		Longitude            float64 `json:"longitude"`
		Timezone             string  `json:"timezone"`
		TimezoneAbbreviation string  `json:"timezone_abbreviation"`
		TimezoneOffset       int     `json:"timezone_offset"`
		Units                string  `json:"units"`
		Currently            Current `json:"currently"`

		Hourly WeatherXuHourly `json:"hourly"`
		Daily  WeatherXuDaily  `json:"daily"`
	} `json:"data"`
}

// ErrorData represents error response from API
type ErrorData struct {
	StatusCode int    `json:"statusCode"`
	Message    string `json:"message"`
}

// Current represents current weather conditions
type Current struct {
	ApparentTemperature float64 `json:"apparentTemperature"`
	CloudCover          float64 `json:"cloudCover"`
	DewPoint            float64 `json:"dewPoint"`
	Humidity            float64 `json:"humidity"`
	Icon                string  `json:"icon"`
	PrecipIntensity     float64 `json:"precipIntensity"`
	Pressure            float64 `json:"pressure"`
	Temperature         float64 `json:"temperature"`
	UvIndex             int     `json:"uvIndex"`
	Visibility          float64 `json:"visibility"`
	WindDirection       float64 `json:"windDirection"`
	WindGust            float64 `json:"windGust"`
	WindSpeed           float64 `json:"windSpeed"`
}

// Hourly represents hourly forecast data
type WeatherXuHourly struct {
	Data []HourlyData `json:"data"`
}

// HourlyData represents weather data for a specific hour
type HourlyData struct {
	ApparentTemperature float64 `json:"apparentTemperature"`
	CloudCover          float64 `json:"cloudCover"`
	DewPoint            float64 `json:"dewPoint"`
	ForecastStart       int64   `json:"forecastStart"`
	Humidity            float64 `json:"humidity"`
	Icon                string  `json:"icon"`
	PrecipIntensity     float64 `json:"precipIntensity"`
	PrecipProbability   float64 `json:"precipProbability"`
	Pressure            float64 `json:"pressure"`
	Temperature         float64 `json:"temperature"`
	UvIndex             int     `json:"uvIndex"`
	Visibility          float64 `json:"visibility"`
	WindDirection       float64 `json:"windDirection"`
	WindGust            float64 `json:"windGust"`
	WindSpeed           float64 `json:"windSpeed"`
}

// Daily represents daily forecast data
type WeatherXuDaily struct {
	Data []DailyData `json:"data"`
}

// DailyData represents weather data for a specific day
type DailyData struct {
	ApparentTemperatureAvg float64 `json:"apparentTemperatureAvg"`
	ApparentTemperatureMax float64 `json:"apparentTemperatureMax"`
	ApparentTemperatureMin float64 `json:"apparentTemperatureMin"`
	CloudCover             float64 `json:"cloudCover"`
	DewPointAvg            float64 `json:"dewPointAvg"`
	DewPointMax            float64 `json:"dewPointMax"`
	DewPointMin            float64 `json:"dewPointMin"`
	ForecastEnd            int64   `json:"forecastEnd"`
	ForecastStart          int64   `json:"forecastStart"`
	Humidity               float64 `json:"humidity"`
	Icon                   string  `json:"icon"`
	MoonPhase              float64 `json:"moonPhase"`
	PrecipIntensity        float64 `json:"precipIntensity"`
	PrecipProbability      float64 `json:"precipProbability"`
	Pressure               float64 `json:"pressure"`
	SunriseTime            int64   `json:"sunriseTime"`
	SunsetTime             int64   `json:"sunsetTime"`
	TemperatureAvg         float64 `json:"temperatureAvg"`
	TemperatureMax         float64 `json:"temperatureMax"`
	TemperatureMin         float64 `json:"temperatureMin"`
	UvIndexMax             int     `json:"uvIndexMax"`
	Visibility             float64 `json:"visibility"`
	WindDirectionAvg       float64 `json:"windDirectionAvg"`
	WindGustAvg            float64 `json:"windGustAvg"`
	WindGustMax            float64 `json:"windGustMax"`
	WindGustMin            float64 `json:"windGustMin"`
	WindSpeedAvg           float64 `json:"windSpeedAvg"`
	WindSpeedMax           float64 `json:"windSpeedMax"`
	WindSpeedMin           float64 `json:"windSpeedMin"`
}

type WeatherXuCodemap struct {
	Code iface.WeatherCode
	Desc string
}

const (
	weatherXuURI = "https://api.weatherxu.com/v1/weather?%s"
)

var (
	weatherXuCodemap = map[string]WeatherXuCodemap{
		"clear":          {iface.CodeSunny, "Clear"},
		"partly_cloudy":  {iface.CodePartlyCloudy, "Partly Cloudy"},
		"mostly_cloudy":  {iface.CodeCloudy, "Cloudy"},
		"cloudy":         {iface.CodeVeryCloudy, "Cloudy"},
		"light_rain":     {iface.CodeLightShowers, "Light Rain"},
		"rain":           {iface.CodeHeavyRain, "Rain"},
		"heavy_rain":     {iface.CodeHeavyRain, "Heavy Rain"},
		"freezing_rain":  {iface.CodeLightSleet, "Freezing Rain"},
		"thunderstorm":   {iface.CodeThunderyHeavyRain, "Thunderstorm"},
		"sleet":          {iface.CodeLightSleet, "Sleet"},
		"light_snow":     {iface.CodeLightSnow, "Light Snow"},
		"snow":           {iface.CodeHeavySnow, "Snow"},
		"heavy_snow":     {iface.CodeHeavySnow, "Heavy Snow"},
		"hail":           {iface.CodeHeavyShowers, "Hail"},
		"windy":          {iface.CodeCloudy, "Windy"},
		"fog":            {iface.CodeFog, "Fog"},
		"mist":           {iface.CodeFog, "Mist"},
		"haze":           {iface.CodeFog, "Haze"},
		"smoke":          {iface.CodeFog, "Smoke"},
		"tornado":        {iface.CodeThunderyHeavyRain, "Tornado"},
		"tropical_storm": {iface.CodeThunderyHeavyRain, "Tropical Storm"},
		"hurricane":      {iface.CodeThunderyHeavyRain, "Hurricane"},
		"sandstorm":      {iface.CodeVeryCloudy, "Sandstorm"},
		"blizzard":       {iface.CodeHeavySnowShowers, "Blizzard"},
	}
)

func (c *weatherXuConfig) Setup() {
	flag.StringVar(&c.apiKey, "weatherxu-api-key", "", "weatherxu backend: the api `KEY` to use")
	flag.StringVar(&c.lang, "weatherxu-lang", "en", "weatherxu backend: the `LANGUAGE` to request from weatherxu")
	flag.BoolVar(&c.debug, "weatherxu-debug", false, "weatherxu backend: print raw requests and responses")
}

func (c *weatherXuConfig) parseDaily(dailyInfo WeatherXuDaily, hourlyInfo WeatherXuHourly) []iface.Day {
	var forecast []iface.Day
	var result []iface.Day
	for _, day := range dailyInfo.Data {
		forecast = append(forecast, c.parseDay(day))
	}
	for _, hourlyData := range hourlyInfo.Data {
		dayIndex := findDailyPeriod(hourlyData.ForecastStart, dailyInfo.Data)
		if dayIndex == -1 {
			continue
		}

		forecast[dayIndex].Slots = append(forecast[dayIndex].Slots, c.parseCurCondHourly(hourlyData))
	}
	for _, day := range forecast {
		if len(day.Slots) > 0 {
			result = append(result, day)
		}
	}
	return result
}

func findDailyPeriod(hourlyTime int64, dailyData []DailyData) int {
	left := 0
	right := len(dailyData) - 1
	for left <= right {
		mid := (left + right) / 2
		if dailyData[mid].ForecastStart <= hourlyTime && hourlyTime <= dailyData[mid].ForecastEnd {
			return mid
		}
		if dailyData[mid].ForecastStart > hourlyTime {
			right = mid - 1
		} else {
			left = mid + 1
		}
	}
	return -1
}
func (c *weatherXuConfig) parseCurCondHourly(hourlyData HourlyData) (ret iface.Cond) {
	ret.Time = time.Unix(hourlyData.ForecastStart, 0)
	ret.Code = iface.CodeUnknown
	if val, ok := weatherXuCodemap[hourlyData.Icon]; ok {
		ret.Code = val.Code
		ret.Desc = val.Desc
	}
	// Convert and set temperature values
	var feelsLike float32 = float32(hourlyData.ApparentTemperature)
	ret.FeelsLikeC = &feelsLike
	var temp float32 = float32(hourlyData.Temperature)
	ret.TempC = &temp

	// Convert and set atmospheric conditions
	var humidity int = int(hourlyData.Humidity * 100)
	ret.Humidity = &humidity
	var visibility float32 = float32(hourlyData.Visibility)
	ret.VisibleDistM = &visibility

	// Add wind information
	var windSpeed float32 = float32(hourlyData.WindSpeed)
	ret.WindspeedKmph = &windSpeed
	var windGust float32 = float32(hourlyData.WindGust)
	ret.WindGustKmph = &windGust
	var windDir int = int(hourlyData.WindDirection)
	ret.WinddirDegree = &windDir

	var precipM float32 = float32(hourlyData.PrecipIntensity)
	ret.PrecipM = &precipM
	var precipProb int = int(hourlyData.PrecipProbability * 100)
	ret.ChanceOfRainPercent = &precipProb

	return ret
}
func (c *weatherXuConfig) parseCurCond(dt int64, current Current) (ret iface.Cond) {
	// Set timestamp
	ret.Time = time.Unix(dt, 0)
	// Map weather code
	ret.Code = iface.CodeUnknown
	if val, ok := weatherXuCodemap[current.Icon]; ok {
		ret.Code = val.Code
		ret.Desc = val.Desc

	}

	// Convert and set temperature values
	var feelsLike float32 = float32(current.ApparentTemperature)
	ret.FeelsLikeC = &feelsLike
	var temp float32 = float32(current.Temperature)
	ret.TempC = &temp

	// Convert and set atmospheric conditions
	var humidity int = int(current.Humidity * 100) // Convert to percentage
	ret.Humidity = &humidity
	var visibility float32 = float32(current.Visibility)
	ret.VisibleDistM = &visibility

	// Add wind information
	var windSpeed float32 = float32(current.WindSpeed)
	ret.WindspeedKmph = &windSpeed
	var windDir int = int(current.WindDirection)
	ret.WinddirDegree = &windDir

	return ret
}

func (c *weatherXuConfig) parseDay(dailyData DailyData) (ret iface.Day) {
	ret.Date = time.Unix(dailyData.ForecastStart, 0)
	ret.Astronomy.Sunrise = time.Unix(dailyData.SunriseTime, 0)
	ret.Astronomy.Sunset = time.Unix(dailyData.SunsetTime, 0)
	return ret
}

func (c *weatherXuConfig) fetch(url string) (*weatherXuResponse, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}

	// Add API key to header
	req.Header.Add("X-API-KEY", c.apiKey)
	if c.debug {
		fmt.Printf("Fetching %s\n", url)
	}

	res, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("unable to get (%s) %v", url, err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("unable to read response body (%s): %v", url, err)
	}

	if c.debug {
		fmt.Printf("Response (%s):\n%s\n", url, string(body))
	}

	var resp weatherXuResponse
	if err = json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("unable to unmarshal response (%s): %v\nThe json body is: %s", url, err, string(body))
	}
	if !resp.Success {
		return nil, fmt.Errorf("error: %s", resp.Error.Message)
	}
	return &resp, nil
}

func (c *weatherXuConfig) Fetch(location string, numdays int) iface.Data {
	var ret iface.Data
	loc := ""

	if len(c.apiKey) == 0 {
		log.Fatal("No weatherxu.com API key specified.\nYou have to register for one at https://weatherxu.com/")
	}
	if matched, err := regexp.MatchString(`^-?[0-9]*(\.[0-9]+)?,-?[0-9]*(\.[0-9]+)?$`, location); !matched || err != nil {
		log.Fatalf("Error: The weatherxu backend only supports latitude,longitude pairs as location %s.\n", location)
	}

	s := strings.Split(location, ",")
	loc = fmt.Sprintf("lat=%s&lon=%s", s[0], s[1])
	requestUrl := fmt.Sprintf(weatherXuURI, loc)
	resp, err := c.fetch(requestUrl)
	if err != nil {
		log.Fatalf("Failed to fetch weather data: %v\n", err)
	}
	ret.Current = c.parseCurCond(resp.Data.Dt, resp.Data.Currently)

	ret.Forecast = c.parseDaily(resp.Data.Daily, resp.Data.Hourly)
	ret.GeoLoc = &iface.LatLon{Latitude: float32(resp.Data.Latitude), Longitude: float32(resp.Data.Longitude)}
	ret.Location = location

	return ret
}

func init() {
	iface.AllBackends["weatherxu"] = &weatherXuConfig{}
}
