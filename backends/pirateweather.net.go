package backends

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"time"

	"github.com/schachmat/wego/iface"
)

const (
	// https://pirateweather.net/en/latest/API
	pirateweatherURI = "https://api.pirateweather.net/forecast"
)

type pirateweatherConfig struct {
	apiKey string
	debug  bool
}

func (c *pirateweatherConfig) Setup() {
	flag.StringVar(&c.apiKey, "pirateweather-api-key", "", "pirateweather backend: the api `KEY` to use")
	flag.BoolVar(&c.debug, "pirateweather-debug", false, "pirateweather backend: print raw req and res")
}

func (c *pirateweatherConfig) Fetch(location string, numdays int) iface.Data {
	if c.apiKey == "" {
		log.Fatal("No pirateweather.net API key specified.\nYou have to register for one at https://pirateweather.net/")
	}

	if c.debug {
		log.Printf("pirateweather location %v", location)
	}

	res := iface.Data{}
	res.Location = location
	reqURI := fmt.Sprintf("%s/%s/%s?extend=hourly&units=si", pirateweatherURI, c.apiKey, location)
	apiRes, err := http.Get(reqURI)
	if err != nil {
		log.Fatalf("Failed to fetch pirateweather data: %v\n", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Println("Failed to close pirateweather response body:", err)
		}
	}(apiRes.Body)

	body, err := io.ReadAll(apiRes.Body)
	if err != nil {
		log.Fatalf("Failed to read pirateweather response: %v\n", err)
	}

	if c.debug {
		safeURI := fmt.Sprintf("%s/REDACTED/%s?extend=hourly&units=si", pirateweatherURI, location)
		log.Println("pirateweather request:", safeURI)
		log.Println("pirateweather status code:", apiRes.StatusCode)
		log.Println("pirateweather response:", string(body))
	}

	weatherData := &pirateweather{}
	if err := json.Unmarshal(body, weatherData); err != nil {
		log.Fatalf("Failed to unmarshal pirateweather response: %v\n", err)
	}

	if c.debug {
		log.Println("hourly datapoints:", len(weatherData.Hourly.Data))
	}

	res.Current = *weatherData.Currently.toCond()
	weatherData.toForecast(&res, numdays)
	latLon := &iface.LatLon{}
	latLon.Latitude = weatherData.Latitude
	latLon.Longitude = weatherData.Longitude
	res.GeoLoc = latLon

	return res
}

type pirateweatherCondFields struct {
	Icon                pirateweatherIcon
	Time                uint
	Summary             string
	Temperature         float32
	ApparentTemperature float32
	PrecipProbability   float32
	PrecipIntensity     float32
	Visibility          float32
	WindSpeed           float32
	WindGust            float32
	WindBearing         float32
	Humidity            float32
}

func parsePirateweatherCond(comp *pirateweatherCondFields) *iface.Cond {
	cond := &iface.Cond{}

	cond.Code = weatherCodeFromPirateweatherIcon(comp.Icon)
	cond.Time = time.Unix(int64(comp.Time), 0)
	cond.Desc = comp.Summary
	cond.TempC = &comp.Temperature
	cond.FeelsLikeC = &comp.ApparentTemperature
	chanceOfRainPercent := int(math.Floor(float64(comp.PrecipProbability * 100.0)))
	cond.ChanceOfRainPercent = &chanceOfRainPercent
	// precipIntensity is in mm/h (SI units); 1 mm = 0.001 m, so divide by 1000 to get m/h
	precipM := comp.PrecipIntensity / 1000.0
	cond.PrecipM = &precipM
	// visibility is in km (SI units); multiply by 1000 to get meters
	visibilityM := comp.Visibility * 1000.0
	cond.VisibleDistM = &visibilityM
	// windSpeed is in m/s (SI units); multiply by 3.6 to get km/h
	windSpeedKmph := comp.WindSpeed * 3.6
	cond.WindspeedKmph = &windSpeedKmph
	windGustKmph := comp.WindGust * 3.6
	cond.WindGustKmph = &windGustKmph
	winddirDegree := int(comp.WindBearing)
	cond.WinddirDegree = &winddirDegree
	humidity := int(comp.Humidity * 100.0)
	cond.Humidity = &humidity

	return cond
}

func (w *pirateweather) toForecast(data *iface.Data, numdays int) {
	var day *iface.Day
	for _, h := range w.Hourly.Data {
		cond := h.toCond()
		if day == nil {
			day = &iface.Day{Date: cond.Time}
		}
		if cond.Time.Year() != day.Date.Year() || cond.Time.Month() != day.Date.Month() || cond.Time.Day() != day.Date.Day() {
			data.Forecast = append(data.Forecast, *day)
			if len(data.Forecast) >= numdays {
				return
			}
			day = &iface.Day{Date: cond.Time}
		}
		day.Slots = append(day.Slots, *cond)
	}
	if day != nil && len(day.Slots) > 0 && len(data.Forecast) < numdays {
		data.Forecast = append(data.Forecast, *day)
	}
}

func weatherCodeFromPirateweatherIcon(weatherIcon pirateweatherIcon) iface.WeatherCode {
	weatherCodeMap := map[pirateweatherIcon]iface.WeatherCode{
		iconClearDay:          iface.CodeSunny,
		iconClearNight:        iface.CodeSunny,
		iconRain:              iface.CodeLightRain,
		iconSnow:              iface.CodeLightSnow,
		iconSleet:             iface.CodeLightSleet,
		iconWind:              iface.CodeCloudy,
		iconFog:               iface.CodeFog,
		iconCloudy:            iface.CodeCloudy,
		iconPartlyCloudyDay:   iface.CodePartlyCloudy,
		iconPartlyCloudyNight: iface.CodePartlyCloudy,
		iconThunderstorm:      iface.CodeThunderyShowers,
		iconHail:              iface.CodeLightSleetShowers,
		iconNone:              iface.CodeUnknown,
	}
	return weatherCodeMap[weatherIcon]
}

type pirateweatherIcon string

const (
	iconClearDay          pirateweatherIcon = "clear-day"
	iconClearNight        pirateweatherIcon = "clear-night"
	iconRain              pirateweatherIcon = "rain"
	iconSnow              pirateweatherIcon = "snow"
	iconSleet             pirateweatherIcon = "sleet"
	iconWind              pirateweatherIcon = "wind"
	iconFog               pirateweatherIcon = "fog"
	iconCloudy            pirateweatherIcon = "cloudy"
	iconPartlyCloudyDay   pirateweatherIcon = "partly-cloudy-day"
	iconPartlyCloudyNight pirateweatherIcon = "partly-cloudy-night"
	iconThunderstorm      pirateweatherIcon = "thunderstorm"
	iconHail              pirateweatherIcon = "hail"
	iconNone              pirateweatherIcon = "none"
)

type pirateweather struct {
	// The requested latitude
	Latitude float32 `json:"latitude"`
	// The requested longitude
	Longitude float32 `json:"longitude"`
	// The requested timezone
	Timezone string `json:"timezone"`
	// The timezone offset in hours
	Offset float32 `json:"offset"`
	// The height above sea level in meters the requested location is
	Elevation int `json:"elevation"`
	// A block containing the current weather information for the requested location
	Currently pirateweatherCurrently `json:"currently"`
	// A block containing the minute-by-minute precipitation intensity for the 60 minutes.
	Minutely struct {
		Summary string                      `json:"summary"`
		Icon    pirateweatherIcon           `json:"icon"`
		Data    []pirateweatherMinutelyData `json:"data"`
	} `json:"minutely"`
	// A block containing the hour-by-hour forecasted conditions for the next 48 hours.
	// If extend hourly is used then the hourly block gives hour-by-hour forecasted
	// conditions for the next 168 hours.
	Hourly struct {
		Summary string            `json:"summary"`
		Icon    pirateweatherIcon `json:"icon"`
		// This should contain 168 blocks of data since we pass 'extend=hourly' in the query params
		Data []pirateweatherHourlyData `json:"data"`
	} `json:"hourly"`
	// A block containing the day-by-day forecasted conditions for the next 7 days.
	Daily struct {
		Summary string                   `json:"summary"`
		Icon    pirateweatherIcon        `json:"icon"`
		Data    []pirateweatherDailyData `json:"data"`
	} `json:"daily"`
	// A block containing miscellaneous data for the API request.
	Flags struct {
		// The models used to generate the forecast.
		Sources []string `json:"sources"`
		// Not implemented, and will always return 0.
		NearestStation float64 `json:"nearest-station"`
		// Indicates which units were used in the forecasts.
		Units string `json:"units"`
		// The version of Pirate Weather used to generate the forecast.
		Version string `json:"version"`
		// The X,Y coordinate and the lat, lon coordinate for the grid cell used for each model used to generate the forecast.
		SourceIDX map[string]struct {
			X    uint    `json:"x"`
			Y    uint    `json:"y"`
			Lat  float32 `json:"lat"`
			Long float32 `json:"long"`
		} `json:"sourceIDX"`
		ProcessTime uint `json:"processTime"`
	} `json:"flags"`
}

type pirateweatherCurrently struct {
	// The time in which the data point begins represented in UNIX time.
	Time uint `json:"time"`
	// A human-readable summary describing the weather conditions for a given data point. The daily summary is calculated between 4:00 am and 4:00 am local time.
	Summary string `json:"summary"`
	// One of a set of icons to provide a visual display of what's happening. The daily icon is calculated between 4:00 am and 4:00 am local time.
	Icon pirateweatherIcon `json:"icon"`
	// The approximate distance to the nearest storm in kilometers.
	NearestStormDistance float32 `json:"nearestStormDistance"`
	// The approximate direction in degrees in which a storm is travelling with 0° representing true north.
	NearestStormBearing float32 `json:"nearestStormBearing"`
	// The rate in which liquid precipitation is falling, in millimeters per hour (SI units).
	PrecipIntensity float32 `json:"precipIntensity"`
	// The probability of precipitation occurring expressed as a decimal between 0 and 1 inclusive.
	PrecipProbability float32 `json:"precipProbability"`
	// The standard deviation of the precipIntensity from the GEFS model.
	PrecipIntensityError float32 `json:"precipIntensityError"`
	// The type of precipitation occurring.
	// If precipIntensity is greater than zero this property will have one of the following values:
	// rain, snow or sleet otherwise the value will be none. sleet is defined as any precipitation which is neither rain nor snow.
	PrecipType string `json:"precipType"`
	// The air temperature in degrees Celsius (SI units).
	Temperature float32 `json:"temperature"`
	// Temperature adjusted for wind and humidity,
	// based the Steadman 1994 approach used by the Australian Bureau of Meteorology.
	// Implemented using the Breezy Weather approach without solar radiation.
	ApparentTemperature float32 `json:"apparentTemperature"`
	// The point in which the air temperature needs (assuming constant pressure) in order to reach a relative humidity of 100%.
	DewPoint float32 `json:"dewPoint"`
	// Relative humidity expressed as a value between 0 and 1 inclusive.
	// This is a percentage of the actual water vapour in the air compared to the total amount of water vapour that can exist at the current temperature.
	Humidity float32 `json:"humidity"`
	// The sea-level pressure in hectopascals (SI units).
	Pressure float32 `json:"pressure"`
	// The current wind speed in meters per second (SI units).
	WindSpeed float32 `json:"windSpeed"`
	// The wind gust in meters per second (SI units).
	WindGust float32 `json:"windGust"`
	// The direction in which the wind is blowing in degrees with 0° representing true north.
	WindBearing float32 `json:"windBearing"`
	// Percentage of the sky that is covered in clouds.
	CloudCover float32 `json:"cloudCover"`
	// The measure of UV radiation as represented as an index starting from 0. 0 to 2 is Low, 3 to 5 is Moderate, 6 and 7 is High, 8 to 10 is Very High and 11+ is considered extreme.
	UVIndex float32 `json:"uvIndex"`
	// The visibility in kilometers (SI units).
	Visibility float32 `json:"visibility"`
	// The density of total atmospheric ozone at a given time in Dobson units.
	Ozone float32 `json:"ozone"`
}

func (c *pirateweatherCurrently) toCond() *iface.Cond {
	condComp := &pirateweatherCondFields{}
	condComp.Icon = c.Icon
	condComp.Time = c.Time
	condComp.Summary = c.Summary
	condComp.Temperature = c.Temperature
	condComp.ApparentTemperature = c.ApparentTemperature
	condComp.PrecipProbability = c.PrecipProbability
	condComp.PrecipIntensity = c.PrecipIntensity
	condComp.Visibility = c.Visibility
	condComp.WindSpeed = c.WindSpeed
	condComp.WindGust = c.WindGust
	condComp.WindBearing = c.WindBearing
	condComp.Humidity = c.Humidity

	return parsePirateweatherCond(condComp)
}

type pirateweatherHourlyData struct {
	// The time in which the data point begins represented in UNIX time.
	Time uint `json:"time"`
	// One of a set of icons to provide a visual display of what's happening. The daily icon is calculated between 4:00 am and 4:00 am local time.
	Icon pirateweatherIcon `json:"icon"`
	// A human-readable summary describing the weather conditions for a given data point. The daily summary is calculated between 4:00 am and 4:00 am local time.
	Summary string `json:"summary"`
	// The rate in which liquid precipitation is falling, in millimeters per hour (SI units).
	PrecipIntensity float32 `json:"precipIntensity"`
	// The probability of precipitation occurring expressed as a decimal between 0 and 1 inclusive.
	PrecipProbability float32 `json:"precipProbability"`
	// The standard deviation of the precipIntensity from the GEFS model.
	PrecipIntensityError float32 `json:"precipIntensityError"`
	// Only on hourly and daily.
	// The total amount of liquid precipitation expected to fall over an hour or a day expressed in centimeters (SI units).
	// For day 0, this is the precipitation during the remaining hours of the day.
	PrecipAccumulation float32 `json:"precipAccumulation"`
	// The type of precipitation occurring.
	// if precipIntensity is greater than zero this property will have one of the following values:
	// rain, snow or sleet otherwise the value will be none. sleet is defined as any precipitation which is neither rain nor snow.
	PrecipType string `json:"precipType"`
	// The air temperature in degrees Celsius (SI units).
	Temperature float32 `json:"temperature"`
	// Temperature adjusted for wind and humidity,
	// based the Steadman 1994 approach used by the Australian Bureau of Meteorology.
	// Implemented using the Breezy Weather approach without solar radiation.
	ApparentTemperature float32 `json:"apparentTemperature"`
	// The point in which the air temperature needs (assuming constant pressure) in order to reach a relative humidity of 100%.
	DewPoint float32 `json:"dewPoint"`
	// Relative humidity expressed as a value between 0 and 1 inclusive.
	// This is a percentage of the actual water vapour in the air compared to the total amount of water vapour that can exist at the current temperature.
	Humidity float32 `json:"humidity"`
	// The sea-level pressure in hectopascals (SI units).
	Pressure float32 `json:"pressure"`
	// The current wind speed in meters per second (SI units).
	WindSpeed float32 `json:"windSpeed"`
	// The wind gust in meters per second (SI units).
	WindGust float32 `json:"windGust"`
	// The direction in which the wind is blowing in degrees with 0° representing true north.
	WindBearing float32 `json:"windBearing"`
	// Percentage of the sky that is covered in clouds.
	CloudCover float32 `json:"cloudCover"`
	// The measure of UV radiation as represented as an index starting from 0. 0 to 2 is Low, 3 to 5 is Moderate, 6 and 7 is High, 8 to 10 is Very High and 11+ is considered extreme.
	UVIndex float32 `json:"uvIndex"`
	// The visibility in kilometers (SI units).
	Visibility float32 `json:"visibility"`
	// The density of total atmospheric ozone at a given time in Dobson units.
	Ozone float32 `json:"ozone"`
}

func (c *pirateweatherHourlyData) toCond() *iface.Cond {
	condComp := &pirateweatherCondFields{}
	condComp.Icon = c.Icon
	condComp.Time = c.Time
	condComp.Summary = c.Summary
	condComp.Temperature = c.Temperature
	condComp.ApparentTemperature = c.ApparentTemperature
	condComp.PrecipProbability = c.PrecipProbability
	condComp.PrecipIntensity = c.PrecipIntensity
	condComp.Visibility = c.Visibility
	condComp.WindSpeed = c.WindSpeed
	condComp.WindGust = c.WindGust
	condComp.WindBearing = c.WindBearing
	condComp.Humidity = c.Humidity

	return parsePirateweatherCond(condComp)
}

type pirateweatherMinutelyData struct {
	// The time in which the data point begins represented in UNIX time.
	Time uint `json:"time"`
	// The rate in which liquid precipitation is falling, in millimeters per hour (SI units).
	PrecipIntensity float32 `json:"precipIntensity"`
	// The probability of precipitation occurring expressed as a decimal between 0 and 1 inclusive.
	PrecipProbability float32 `json:"precipProbability"`
	// The standard deviation of the precipIntensity from the GEFS model.
	PrecipIntensityError float32 `json:"precipIntensityError"`
	// The type of precipitation occurring.
	// if precipIntensity is greater than zero this property will have one of the following values:
	// rain, snow or sleet otherwise the value will be none. sleet is defined as any precipitation which is neither rain nor snow.
	PrecipType string `json:"precipType"`
}

type pirateweatherDailyData struct {
	// The time in which the data point begins represented in UNIX time.
	Time uint `json:"time"`
	// One of a set of icons to provide a visual display of what's happening. The daily icon is calculated between 4:00 am and 4:00 am local time.
	Icon string `json:"icon"`
	// A human-readable summary describing the weather conditions for a given data point. The daily summary is calculated between 4:00 am and 4:00 am local time.
	Summary string `json:"summary"`
	// Only on daily. The time when the sun rises for a given day represented in UNIX time.
	SunriseTime uint `json:"sunriseTime"`
	// Only on daily. The time when the sun sets for a given day represented in UNIX time.
	SunsetTime uint `json:"sunsetTime"`
	// Only on daily.
	// The fractional lunation number for the given day. 0.00 represents a new moon, 0.25 represents the first quarter, 0.50 represents a full moon and 0.75 represents the last quarter.
	MoonPhase float32 `json:"moonPhase"`
	// The rate in which liquid precipitation is falling, in millimeters per hour (SI units).
	PrecipIntensity float32 `json:"precipIntensity"`
	// Only on daily. The maximum value of precipIntensity for the given day.
	PrecipIntensityMax float32 `json:"precipIntensityMax"`
	// Only on daily. The point in which the maximum precipIntensity occurs represented in UNIX time.
	PrecipIntensityMaxTime uint `json:"precipIntensityMaxTime"`
	// The probability of precipitation occurring expressed as a decimal between 0 and 1 inclusive.
	PrecipProbability float32 `json:"precipProbability"`
	// Only on hourly and daily.
	// The total amount of liquid precipitation expected to fall over an hour or a day expressed in centimeters (SI units).
	// For day 0, this is the precipitation during the remaining hours of the day.
	PrecipAccumulation float32 `json:"precipAccumulation"`
	// The type of precipitation occurring.
	// if precipIntensity is greater than zero this property will have one of the following values:
	// rain, snow or sleet otherwise the value will be none. sleet is defined as any precipitation which is neither rain nor snow.
	PrecipType string `json:"precipType"`
	// Only on daily. The daytime high temperature calculated between 6:01 am and 6:00 pm local time.
	TemperatureHigh float32 `json:"temperatureHigh"`
	// Only on daily. The time in which the high temperature occurs represented in UNIX time.
	TemperatureHighTime uint `json:"temperatureHighTime"`
	// Only on daily. The overnight low temperature calculated between 6:01 pm and 6:00 am local time.
	TemperatureLow float32 `json:"temperatureLow"`
	// Only on daily. The time in which the low temperature occurs represented in UNIX time.
	TemperatureLowTime uint `json:"temperatureLowTime"`
	// Only on daily.
	// The maximum "feels like" temperature during the daytime, from 6:00 am to 6:00 pm.
	ApparentTemperatureHigh float32 `json:"apparentTemperatureHigh"`
	// Only on daily.
	// The time of the maximum "feels like" temperature during the daytime,
	// from 6:00 am to 6:00 pm.
	ApparentTemperatureHighTime uint `json:"apparentTemperatureHighTime"`
	// Only on daily.
	// The minimum "feels like" temperature during the daytime, from 6:00 am to 6:00 pm.
	ApparentTemperatureLow float32 `json:"apparentTemperatureLow"`
	// Only on daily.
	// The time of the minimum "feels like" temperature during the daytime,
	// from 6:00 am to 6:00 pm.
	ApparentTemperatureLowTime float32 `json:"apparentTemperatureLowTime"`
	// The sea-level pressure in hectopascals (SI units).
	Pressure float32 `json:"pressure"`
	// The wind speed in meters per second (SI units).
	WindSpeed float32 `json:"windSpeed"`
	// The wind gust in meters per second (SI units).
	WindGust float32 `json:"windGust"`
	// Only on daily. The time in which the maximum wind gust occurs during the day represented in UNIX time.
	WindGustTime uint `json:"windGustTime"`
	// The direction in which the wind is blowing in degrees with 0° representing true north.
	WindBearing float32 `json:"windBearing"`
	// Percentage of the sky that is covered in clouds.
	// This value will be between 0 and 1 inclusive.
	// Calculated from the the GFS (#650) or HRRR (#115) TCDC variable for the entire atmosphere.
	CloudCover float32 `json:"cloudCover"`
	// The measure of UV radiation as represented as an index starting from 0. 0 to 2 is Low, 3 to 5 is Moderate, 6 and 7 is High, 8 to 10 is Very High and 11+ is considered extreme.
	UVIndex float32 `json:"uvIndex"`
	// Only on daily. The time in which the maximum uvIndex occurs during the day.
	UVIndexTime uint `json:"uvIndexTime"`
	// The visibility in kilometers (SI units).
	Visibility float32 `json:"visibility"`
	// Only on daily. The minimum temperature calculated between 12:00 am and 11:59 pm local time.
	TemperatureMin float32 `json:"temperatureMin"`
	// Only on daily. The time in which the minimum temperature occurs represented in UNIX time.
	TemperatureMinTime uint `json:"temperatureMinTime"`
	// Only on daily. The maximum temperature calculated between 12:00 am and 11:59 pm local time.
	TemperatureMax float32 `json:"temperatureMax"`
	// Only on daily. The time in which the maximum temperature occurs represented in UNIX time.
	TemperatureMaxTime uint `json:"temperatureMaxTime"`
	// Only on daily.
	// The minimum "feels like" temperature during a day, from 12:00 am and 11:59 pm.
	ApparentTemperatureMin float32 `json:"apparentTemperatureMin"`
	// Only on daily.
	// The time (in UTC) that the minimum "feels like" temperature occurs during a day,
	// from 12:00 am and 11:59 pm.
	ApparentTemperatureMinTime float32 `json:"apparentTemperatureMinTime"`
	// Only on daily.
	// The maximum "feels like" temperature during a day, from midnight to midnight.
	ApparentTemperatureMax float32 `json:"apparentTemperatureMax"`
	// Only on daily.
	// The time (in UTC) that the maximum "feels like" temperature occurs during a day,
	// from 12:00 am and 11:59 pm.
	ApparentTemperatureMaxTime uint `json:"apparentTemperatureMaxTime"`
}

func init() {
	iface.AllBackends["pirateweather.net"] = &pirateweatherConfig{}
}
