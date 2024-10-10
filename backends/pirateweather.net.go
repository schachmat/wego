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

type PirateweatherConfig struct {
	apiKey string
	debug  bool
}

func (c *PirateweatherConfig) Setup() {
	flag.StringVar(&c.apiKey, "pirateweather-api-key", "", "pirateweather backend: the api `KEY` to use")
	flag.BoolVar(&c.debug, "pirateweather-debug", false, "pirateweather backend: print raw req and res")
}

func (c *PirateweatherConfig) Fetch(location string, numdays int) iface.Data {
	if c.debug {
		log.Printf("pirateweather location %v", location)
	}

	res := iface.Data{}
	reqURI := fmt.Sprintf("%s/%s/%s?extend=hourly&units=si", pirateweatherURI, c.apiKey, location)
	apiRes, err := http.Get(reqURI)
	if err != nil {
		panic(err)
	}
	defer apiRes.Body.Close()

	body, err := io.ReadAll(apiRes.Body)
	if err != nil {
		panic(err)
	}

	if c.debug {
		log.Println("pirateweather request:", reqURI)
		log.Println("pirateweather status code:", apiRes.StatusCode)
		data, _ := json.MarshalIndent(body, "", "\t")
		log.Println("pirateweather response:", string(data))
	}

	weatherData := &Pirateweather{}
	if err := json.Unmarshal(body, weatherData); err != nil {
		panic(err)
	}

	if c.debug {
		log.Println("pirateweather data", weatherData)
		log.Println("hourly datapoints:", len(weatherData.Hourly.Data))
	}

	res.Current = *weatherData.Currently.toCond()
	weatherData.toForecast(&res)

	return res
}

type CondCompatible struct {
	Icon                PirateweatherIcon
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

func parseCond(comp *CondCompatible) *iface.Cond {
	cond := &iface.Cond{}

	cond.Code = weatherCodeFromPirateweatherIcon(comp.Icon)
	cond.Time = time.Unix(int64(comp.Time), 0)
	cond.Desc = comp.Summary
	cond.TempC = &comp.Temperature
	cond.FeelsLikeC = &comp.ApparentTemperature
	chanceOfRainPercent := int(math.Floor(float64(comp.PrecipProbability * 100.0)))
	cond.ChanceOfRainPercent = &chanceOfRainPercent
	precipM := comp.PrecipIntensity * 1000.0
	cond.PrecipM = &precipM
	visibilityDistM := comp.Visibility / 1000.0
	cond.VisibleDistM = &visibilityDistM
	cond.WindspeedKmph = &comp.WindSpeed
	cond.WindGustKmph = &comp.WindGust
	winddirDegree := int(comp.WindBearing)
	cond.WinddirDegree = &winddirDegree
	humidity := int(comp.Humidity * 100.0)
	cond.Humidity = &humidity

	return cond
}

func (w *Pirateweather) toForecast(data *iface.Data) {
	day1 := &iface.Day{}
	day1.Slots = parseHourlyDataForSingleDay(0, 6, w.Hourly.Data)
	day1.Date = day1.Slots[0].Time
	data.Forecast = append(data.Forecast, *day1)

	day2 := &iface.Day{}
	day2.Slots = parseHourlyDataForSingleDay(7, 14, w.Hourly.Data)
	day2.Date = day2.Slots[0].Time
	data.Forecast = append(data.Forecast, *day2)

	day3 := &iface.Day{}
	day3.Slots = parseHourlyDataForSingleDay(15, 22, w.Hourly.Data)
	day3.Date = day3.Slots[0].Time
	data.Forecast = append(data.Forecast, *day3)

	day4 := &iface.Day{}
	day4.Slots = parseHourlyDataForSingleDay(23, 30, w.Hourly.Data)
	day4.Date = day4.Slots[0].Time
	data.Forecast = append(data.Forecast, *day4)

	day5 := &iface.Day{}
	day5.Slots = parseHourlyDataForSingleDay(31, 38, w.Hourly.Data)
	day5.Date = day5.Slots[0].Time
	data.Forecast = append(data.Forecast, *day5)

	day6 := &iface.Day{}
	day6.Slots = parseHourlyDataForSingleDay(39, 46, w.Hourly.Data)
	day6.Date = day6.Slots[0].Time
	data.Forecast = append(data.Forecast, *day6)

	day7 := &iface.Day{}
	day7.Slots = parseHourlyDataForSingleDay(47, 54, w.Hourly.Data)
	day7.Date = day7.Slots[0].Time
	data.Forecast = append(data.Forecast, *day7)
}

func weatherCodeFromPirateweatherIcon(weatherIcon PirateweatherIcon) iface.WeatherCode {
	weatherCodeMap := map[PirateweatherIcon]iface.WeatherCode{
		IconClearDay:          iface.CodeSunny,
		IconClearNight:        iface.CodeSunny,
		IconRain:              iface.CodeLightRain,
		IconSnow:              iface.CodeLightSnow,
		IconSleet:             iface.CodeLightSleet,
		IconWind:              iface.CodeCloudy,
		IconFog:               iface.CodeFog,
		IconCloudy:            iface.CodeCloudy,
		IconPartlyCloudyDay:   iface.CodePartlyCloudy,
		IconPartlyCloudyNight: iface.CodePartlyCloudy,
		IconThunderstorm:      iface.CodeThunderyShowers,
		IconHail:              iface.CodeLightSleetShowers,
		IconNone:              iface.CodeUnknown,
	}
	return weatherCodeMap[weatherIcon]
}

type PirateweatherIcon string

const (
	IconClearDay          PirateweatherIcon = "clear-day"
	IconClearNight        PirateweatherIcon = "clear-night"
	IconRain              PirateweatherIcon = "rain"
	IconSnow              PirateweatherIcon = "snow"
	IconSleet             PirateweatherIcon = "sleet"
	IconWind              PirateweatherIcon = "wind"
	IconFog               PirateweatherIcon = "fog"
	IconCloudy            PirateweatherIcon = "cloudy"
	IconPartlyCloudyDay   PirateweatherIcon = "partly-cloudy-day"
	IconPartlyCloudyNight PirateweatherIcon = "partly-cloudy-night"
	IconThunderstorm      PirateweatherIcon = "thunderstorm"
	IconHail              PirateweatherIcon = "hail"
	IconNone              PirateweatherIcon = "none"
)

type Pirateweather struct {
	// The requested latitude
	Latitude float32 `json:"latitude"`
	// The requested longtitude
	Longitude float32 `json:"longitude"`
	// The requested timezone
	Timezone string `json:"timezone"`
	// The timezone offset in hours
	Offset float32 `json:"offset"`
	// The height above sea level in meters the requested location is
	Elevation uint `json:"elevation"`
	// A block containing the current weather information for the requested location
	Currently PirateweatherCurrently `json:"currently"`
	// A block containing the minute-by-minute precipitation intensity for the 60 minutes.
	Minutely struct {
		Summary string                      `json:"summary"`
		Icon    PirateweatherIcon           `json:"icon"`
		Data    []PirateweatherMinutelyData `json:"data"`
	} `json:"minutely"`
	// A block containing the hour-by-hour forcasted conditions for the next 48 hours.
	// If extend hourly is used then the hourly block gives hour-by-hour forecasted
	// conditions for the next 168 hours.
	Hourly struct {
		Summary string            `json:"summary"`
		Icon    PirateweatherIcon `json:"icon"`
		// This should contain 168 blocks of data since we pass 'extend=hourly' in the query params
		Data []PirateweatherHourlyData `json:"data"`
	} `json:"hourly"`
	// A block containing the day-by-day forecasted conditions for the next 7 days.
	Daily struct {
		Summary string                   `json:"summary"`
		Icon    PirateweatherIcon        `json:"icon"`
		Data    []PirateweatherDailyData `json:"data"`
	} `json:"daily"`
	// A block containing miscellaneous data for the API request.
	Flags struct {
		// The models used to generate the forecast.
		Sources []string `json:"sources"`
		// Not implemented, and will always return 0.
		NearestStation uint `json:"nearest-station"`
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

type PirateweatherCurrently struct {
	// The time in which the data point begins represented in UNIX time.
	Time uint `json:"time"`
	// A human-readable summary describing the weather conditions for a given data point. The daily summary is calculated between 4:00 am and 4:00 am local time.
	Summary string `json:"summary"`
	// One of a set of icons to provide a visual display of what's happening. The daily icon is calculated between 4:00 am and 4:00 am local time.
	Icon PirateweatherIcon `json:"icon"`
	// The approximate distance to the nearest storm in kilometers or miles depending on the requested units.
	NearestStormDistance float32 `json:"nearestStormDistance"`
	// The approximate direction in degrees in which a storm is travelling with 0° representing true north.
	NearestStormBearing float32 `json:"nearestStormBearing"`
	// The rate in which liquid precipitation is falling. This value is expressed in millimetres per hour or inches per hour depending on the requested units.
	PrecipIntensity float32 `json:"precipIntensity"`
	// The probability of precipitation occurring expressed as a decimal between 0 and 1 inclusive.
	PrecipProbability float32 `json:"precipProbability"`
	// The standard deviation of the precipIntensity from the GEFS model.
	PrecipIntensityError float32 `json:"precipIntensityError"`
	// The type of precipitation occurring.
	// If precipIntensity is greater than zero this property will have one of the following values:
	// rain, snow or sleet otherwise the value will be none. sleet is defined as any precipitation which is neither rain nor snow.
	PrecipType string `json:"precipType"`
	// The air temperature in degrees Celsius or degrees Fahrenheit depending on the requested units
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
	// The sea-level pressure represented in hectopascals or millibars depending on the requested units.
	Pressure float32 `json:"pressure"`
	// The current wind speed in kilometres per hour or miles per hour depending on the requested units.
	WindSpeed float32 `json:"windSpeed"`
	// The wind gust in kilometres per hour or miles per hour depending on the requested units.
	WindGust float32 `json:"windGust"`
	// The direction in which the wind is blowing in degrees with 0° representing true north.
	WindBearing float32 `json:"windBearing"`
	// Percentage of the sky that is covered in clouds.
	CloudCover float32 `json:"cloudCover"`
	// The measure of UV radiation as represented as an index starting from 0. 0 to 2 is Low, 3 to 5 is Moderate, 6 and 7 is High, 8 to 10 is Very High and 11+ is considered extreme.
	UVIndex float32 `json:"uvIndex"`
	// The visibility in kilometres or miles depending on the requested units.
	Visibility float32 `json:"visibility"`
	// The density of total atmospheric ozone at a given time in Dobson units.
	Ozone float32 `json:"ozone"`
}

func (c *PirateweatherCurrently) toCond() *iface.Cond {
	condComp := &CondCompatible{}
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

	return parseCond(condComp)
}

type PirateweatherMinutelyData struct {
	// The time in which the data point begins represented in UNIX time.
	Time uint `json:"time"`
	// The rate in which liquid precipitation is falling. This value is expressed in millimetres per hour or inches per hour depending on the requested units.
	PrecipIntensity float32 `json:"precipIntensity"`
	// The probability of precipitation occurring expressed as a decimal between 0 and 1 inclusive.
	PrecipProbability float32 `json:"precipProbability"`
	// The standard deviation of the precipIntensity from the GEFS model.
	PrecipIntensityError float32 `json:"precipIntensityError"`
	// The type of precipitation occurring.
	// if precipintensity is greater than zero this property will have one of the following values:
	// rain, snow or sleet otherwise the value will be none. sleet is defined as any precipitation which is neither rain nor snow.
	PrecipType string `json:"precipType"`
}

type PirateweatherHourlyData struct {
	// The time in which the data point begins represented in UNIX time.
	Time uint `json:"time"`
	// One of a set of icons to provide a visual display of what's happening. The daily icon is calculated between 4:00 am and 4:00 am local time.
	Icon PirateweatherIcon `json:"icon"`
	// A human-readable summary describing the weather conditions for a given data point. The daily summary is calculated between 4:00 am and 4:00 am local time.
	Summary string `json:"summary"`
	// The rate in which liquid precipitation is falling. This value is expressed in millimetres per hour or inches per hour depending on the requested units.
	PrecipIntensity float32 `json:"precipIntensity"`
	// The probability of precipitation occurring expressed as a decimal between 0 and 1 inclusive.
	PrecipProbability float32 `json:"precipProbability"`
	// The standard deviation of the precipIntensity from the GEFS model.
	PrecipIntensityError float32 `json:"precipIntensityError"`
	// Only on hourly and daily.
	// The total amount of liquid precipitation expected to fall over an hour or a day expressed in centimetres or inches depending on the requested units.
	// For day 0, this is the precipitation during the remaining hours of the day.
	PrecipAccumulation float32 `json:"precipAccumulation"`
	// The type of precipitation occurring.
	// if precipintensity is greater than zero this property will have one of the following values:
	// rain, snow or sleet otherwise the value will be none. sleet is defined as any precipitation which is neither rain nor snow.
	PrecipType string `json:"precipType"`
	// The air temperature in degrees Celsius or degrees Fahrenheit depending on the requested units
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
	// The sea-level pressure represented in hectopascals or millibars depending on the requested units.
	Pressure float32 `json:"pressure"`
	// The current wind speed in kilometres per hour or miles per hour depending on the requested units.
	WindSpeed float32 `json:"windSpeed"`
	// The wind gust in kilometres per hour or miles per hour depending on the requested units.
	WindGust float32 `json:"windGust"`
	// The direction in which the wind is blowing in degrees with 0° representing true north.
	WindBearing float32 `json:"windBearing"`
	// Percentage of the sky that is covered in clouds.
	CloudCover float32 `json:"cloudCover"`
	// The measure of UV radiation as represented as an index starting from 0. 0 to 2 is Low, 3 to 5 is Moderate, 6 and 7 is High, 8 to 10 is Very High and 11+ is considered extreme.
	UVIndex float32 `json:"uvIndex"`
	// The visibility in kilometres or miles depending on the requested units.
	Visibility float32 `json:"visibility"`
	// The density of total atmospheric ozone at a given time in Dobson units.
	Ozone float32 `json:"ozone"`
}

func (c *PirateweatherHourlyData) toCond() *iface.Cond {
	condComp := &CondCompatible{}
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

	return parseCond(condComp)
}

func parseHourlyDataForSingleDay(lowerIndex, upperIndex int, data []PirateweatherHourlyData) []iface.Cond {
	condArr := []iface.Cond{}
	for i := lowerIndex; i < upperIndex; i++ {
		a := data[i]
		condArr = append(condArr, *a.toCond())
	}

	return condArr
}

type PirateweatherDailyData struct {
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
	// The rate in which liquid precipitation is falling. This value is expressed in millimetres per hour or inches per hour depending on the requested units.
	PrecipIntensity float32 `json:"precipIntensity"`
	// Only on daily. The maximum value of precipIntensity for the given day.
	PrecipIntensityMax float32 `json:"precipIntensityMax"`
	// Only on daily. The point in which the maximum precipIntensity occurs represented in UNIX time.
	PrecipIntensityMaxTime uint `json:"precipIntensityMaxTime"`
	// The probability of precipitation occurring expressed as a decimal between 0 and 1 inclusive.
	PrecipProbability float32 `json:"precipProbability"`
	// Only on hourly and daily.
	// The total amount of liquid precipitation expected to fall over an hour or a day expressed in centimetres or inches depending on the requested units.
	// For day 0, this is the precipitation during the remaining hours of the day.
	PrecipAccumulation float32 `json:"precipAccumulation"`
	// The type of precipitation occurring.
	// if precipintensity is greater than zero this property will have one of the following values:
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
	// The sea-level pressure represented in hectopascals or millibars depending on the requested units.
	Pressure float32 `json:"pressure"`
	// The current wind speed in kilometres per hour or miles per hour depending on the requested units.
	WindSpeed float32 `json:"windSpeed"`
	// The wind gust in kilometres per hour or miles per hour depending on the requested units.
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
	// The visibility in kilometres or miles depending on the requested units.
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
	// The minimum "feels like" temperature during a day, from from 12:00 am and 11:59 pm.
	ApparentTemperatureMin float32 `json:"apparentTemperatureMin"`
	// Only on daily.
	// The time (in UTC) that the minimum "feels like" temperature occurs during a day,
	// from from 12:00 am and 11:59 pm.
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
	iface.AllBackends["pirateweather.net"] = &PirateweatherConfig{}
}
