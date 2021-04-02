package backends

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/schachmat/wego/iface"
	"golang.org/x/net/html/charset"
)

type canadianWeatherConfig struct {
	debug bool
}

type canadianResponse struct {
	XMLName                   xml.Name `xml:"siteData"`
	Text                      string   `xml:",chardata"`
	Xsi                       string   `xml:"xsi,attr"`
	NoNamespaceSchemaLocation string   `xml:"noNamespaceSchemaLocation,attr"`
	License                   string   `xml:"license"`
	DateTime                  []struct {
		Text      string `xml:",chardata"`
		Name      string `xml:"name,attr"`
		Zone      string `xml:"zone,attr"`
		UTCOffset string `xml:"UTCOffset,attr"`
		Year      string `xml:"year"`
		Month     struct {
			Text string `xml:",chardata"`
			Name string `xml:"name,attr"`
		} `xml:"month"`
		Day struct {
			Text string `xml:",chardata"`
			Name string `xml:"name,attr"`
		} `xml:"day"`
		Hour        string `xml:"hour"`
		Minute      string `xml:"minute"`
		TimeStamp   string `xml:"timeStamp"`
		TextSummary string `xml:"textSummary"`
	} `xml:"dateTime"`
	Location struct {
		Text      string `xml:",chardata"`
		Continent string `xml:"continent"`
		Country   struct {
			Text string `xml:",chardata"`
			Code string `xml:"code,attr"`
		} `xml:"country"`
		Province struct {
			Text string `xml:",chardata"`
			Code string `xml:"code,attr"`
		} `xml:"province"`
		Name struct {
			Text string `xml:",chardata"`
			Code string `xml:"code,attr"`
			Lat  string `xml:"lat,attr"`
			Lon  string `xml:"lon,attr"`
		} `xml:"name"`
		Region string `xml:"region"`
	} `xml:"location"`
	Warnings          string            `xml:"warnings"`
	CurrentConditions CurrentConditions `xml:"currentConditions"`
	ForecastGroup     struct {
		Text     string `xml:",chardata"`
		DateTime []struct {
			Text      string `xml:",chardata"`
			Name      string `xml:"name,attr"`
			Zone      string `xml:"zone,attr"`
			UTCOffset string `xml:"UTCOffset,attr"`
			Year      int    `xml:"year"`
			Month     struct {
				Value int    `xml:",chardata"`
				Name  string `xml:"name,attr"`
			} `xml:"month"`
			Day struct {
				Value int    `xml:",chardata"`
				Name  string `xml:"name,attr"`
			} `xml:"day"`
			Hour        int    `xml:"hour"`
			Minute      int    `xml:"minute"`
			TimeStamp   string `xml:"timeStamp"`
			TextSummary string `xml:"textSummary"`
		} `xml:"dateTime"`
		RegionalNormals struct {
			Text        string `xml:",chardata"`
			TextSummary string `xml:"textSummary"`
			Temperature []struct {
				Text     string `xml:",chardata"`
				UnitType string `xml:"unitType,attr"`
				Units    string `xml:"units,attr"`
				Class    string `xml:"class,attr"`
			} `xml:"temperature"`
		} `xml:"regionalNormals"`
		Forecast []Forecast `xml:"forecast"`
	} `xml:"forecastGroup"`
	HourlyForecastGroup struct {
		Text     string `xml:",chardata"`
		DateTime []struct {
			Text      string `xml:",chardata"`
			Name      string `xml:"name,attr"`
			Zone      string `xml:"zone,attr"`
			UTCOffset string `xml:"UTCOffset,attr"`
			Year      string `xml:"year"`
			Month     struct {
				Text string `xml:",chardata"`
				Name string `xml:"name,attr"`
			} `xml:"month"`
			Day struct {
				Text string `xml:",chardata"`
				Name string `xml:"name,attr"`
			} `xml:"day"`
			Hour        string `xml:"hour"`
			Minute      string `xml:"minute"`
			TimeStamp   string `xml:"timeStamp"`
			TextSummary string `xml:"textSummary"`
		} `xml:"dateTime"`
		HourlyForecast []HourlyForecast `xml:"hourlyForecast"`
	} `xml:"hourlyForecastGroup"`
	YesterdayConditions struct {
		Text        string `xml:",chardata"`
		Temperature []struct {
			Text     string `xml:",chardata"`
			UnitType string `xml:"unitType,attr"`
			Units    string `xml:"units,attr"`
			Class    string `xml:"class,attr"`
		} `xml:"temperature"`
		Precip struct {
			Text     string `xml:",chardata"`
			UnitType string `xml:"unitType,attr"`
			Units    string `xml:"units,attr"`
		} `xml:"precip"`
	} `xml:"yesterdayConditions"`
	RiseSet struct {
		Text       string `xml:",chardata"`
		Disclaimer string `xml:"disclaimer"`
		DateTime   []struct {
			Text      string `xml:",chardata"`
			Name      string `xml:"name,attr"`
			Zone      string `xml:"zone,attr"`
			UTCOffset string `xml:"UTCOffset,attr"`
			Year      string `xml:"year"`
			Month     struct {
				Text string `xml:",chardata"`
				Name string `xml:"name,attr"`
			} `xml:"month"`
			Day struct {
				Text string `xml:",chardata"`
				Name string `xml:"name,attr"`
			} `xml:"day"`
			Hour        string `xml:"hour"`
			Minute      string `xml:"minute"`
			TimeStamp   string `xml:"timeStamp"`
			TextSummary string `xml:"textSummary"`
		} `xml:"dateTime"`
	} `xml:"riseSet"`
	Almanac struct {
		Text        string `xml:",chardata"`
		Temperature []struct {
			Text     string `xml:",chardata"`
			Class    string `xml:"class,attr"`
			Period   string `xml:"period,attr"`
			UnitType string `xml:"unitType,attr"`
			Units    string `xml:"units,attr"`
			Year     string `xml:"year,attr"`
		} `xml:"temperature"`
		Precipitation []struct {
			Text     string `xml:",chardata"`
			Class    string `xml:"class,attr"`
			Period   string `xml:"period,attr"`
			UnitType string `xml:"unitType,attr"`
			Units    string `xml:"units,attr"`
			Year     string `xml:"year,attr"`
		} `xml:"precipitation"`
		Pop struct {
			Text  string `xml:",chardata"`
			Units string `xml:"units,attr"`
		} `xml:"pop"`
	} `xml:"almanac"`
}

type CurrentConditions struct {
	Text    string `xml:",chardata"`
	Station struct {
		Text string `xml:",chardata"`
		Code string `xml:"code,attr"`
		Lat  string `xml:"lat,attr"`
		Lon  string `xml:"lon,attr"`
	} `xml:"station"`
	DateTime []struct {
		Text      string `xml:",chardata"`
		Name      string `xml:"name,attr"`
		Zone      string `xml:"zone,attr"`
		UTCOffset string `xml:"UTCOffset,attr"`
		Year      string `xml:"year"`
		Month     struct {
			Text string `xml:",chardata"`
			Name string `xml:"name,attr"`
		} `xml:"month"`
		Day struct {
			Text string `xml:",chardata"`
			Name string `xml:"name,attr"`
		} `xml:"day"`
		Hour        string `xml:"hour"`
		Minute      string `xml:"minute"`
		TimeStamp   string `xml:"timeStamp"`
		TextSummary string `xml:"textSummary"`
	} `xml:"dateTime"`
	Condition string `xml:"condition"`
	IconCode  struct {
		Text   string `xml:",chardata"`
		Format string `xml:"format,attr"`
	} `xml:"iconCode"`
	Temperature struct {
		Value    float32 `xml:",chardata"`
		UnitType string  `xml:"unitType,attr"`
		Units    string  `xml:"units,attr"`
	} `xml:"temperature"`
	Dewpoint struct {
		Text     string `xml:",chardata"`
		UnitType string `xml:"unitType,attr"`
		Units    string `xml:"units,attr"`
	} `xml:"dewpoint"`
	WindChill struct {
		Text     string `xml:",chardata"`
		UnitType string `xml:"unitType,attr"`
	} `xml:"windChill"`
	Pressure struct {
		Text     string `xml:",chardata"`
		UnitType string `xml:"unitType,attr"`
		Units    string `xml:"units,attr"`
		Change   string `xml:"change,attr"`
		Tendency string `xml:"tendency,attr"`
	} `xml:"pressure"`
	Visibility struct {
		Text     string `xml:",chardata"`
		UnitType string `xml:"unitType,attr"`
		Units    string `xml:"units,attr"`
	} `xml:"visibility"`
	RelativeHumidity struct {
		Text  string `xml:",chardata"`
		Units string `xml:"units,attr"`
	} `xml:"relativeHumidity"`
	Wind struct {
		Text  string `xml:",chardata"`
		Speed struct {
			Text     string `xml:",chardata"`
			UnitType string `xml:"unitType,attr"`
			Units    string `xml:"units,attr"`
		} `xml:"speed"`
		Gust struct {
			Text     string `xml:",chardata"`
			UnitType string `xml:"unitType,attr"`
			Units    string `xml:"units,attr"`
		} `xml:"gust"`
		Direction string `xml:"direction"`
		Bearing   struct {
			Text  string `xml:",chardata"`
			Units string `xml:"units,attr"`
		} `xml:"bearing"`
	} `xml:"wind"`
}

type Forecast struct {
	Text   string `xml:",chardata"`
	Period struct {
		Text             string `xml:",chardata"`
		TextForecastName string `xml:"textForecastName,attr"`
	} `xml:"period"`
	TextSummary string `xml:"textSummary"`
	CloudPrecip struct {
		Text        string `xml:",chardata"`
		TextSummary string `xml:"textSummary"`
	} `xml:"cloudPrecip"`
	AbbreviatedForecast struct {
		Text     string `xml:",chardata"`
		IconCode struct {
			Text   string `xml:",chardata"`
			Format string `xml:"format,attr"`
		} `xml:"iconCode"`
		Pop struct {
			Text  string `xml:",chardata"`
			Units string `xml:"units,attr"`
		} `xml:"pop"`
		TextSummary string `xml:"textSummary"`
	} `xml:"abbreviatedForecast"`
	Temperatures struct {
		Text        string `xml:",chardata"`
		TextSummary string `xml:"textSummary"`
		Temperature struct {
			Value    float32 `xml:",chardata"`
			UnitType string  `xml:"unitType,attr"`
			Units    string  `xml:"units,attr"`
			Class    string  `xml:"class,attr"`
		} `xml:"temperature"`
	} `xml:"temperatures"`
	Winds struct {
		Text        string `xml:",chardata"`
		TextSummary string `xml:"textSummary"`
		Wind        []struct {
			Text  string `xml:",chardata"`
			Index string `xml:"index,attr"`
			Rank  string `xml:"rank,attr"`
			Speed struct {
				Text     string `xml:",chardata"`
				UnitType string `xml:"unitType,attr"`
				Units    string `xml:"units,attr"`
			} `xml:"speed"`
			Gust struct {
				Text     string `xml:",chardata"`
				UnitType string `xml:"unitType,attr"`
				Units    string `xml:"units,attr"`
			} `xml:"gust"`
			Direction string `xml:"direction"`
			Bearing   struct {
				Text  string `xml:",chardata"`
				Units string `xml:"units,attr"`
			} `xml:"bearing"`
		} `xml:"wind"`
	} `xml:"winds"`
	Humidex       string `xml:"humidex"`
	Precipitation struct {
		Text        string `xml:",chardata"`
		TextSummary string `xml:"textSummary"`
		PrecipType  struct {
			Text  string `xml:",chardata"`
			Start string `xml:"start,attr"`
			End   string `xml:"end,attr"`
		} `xml:"precipType"`
	} `xml:"precipitation"`
	WindChill struct {
		Text        string `xml:",chardata"`
		TextSummary string `xml:"textSummary"`
		Calculated  struct {
			Text     string `xml:",chardata"`
			UnitType string `xml:"unitType,attr"`
			Class    string `xml:"class,attr"`
		} `xml:"calculated"`
		Frostbite string `xml:"frostbite"`
	} `xml:"windChill"`
	Uv struct {
		Text        string `xml:",chardata"`
		Category    string `xml:"category,attr"`
		Index       string `xml:"index"`
		TextSummary string `xml:"textSummary"`
	} `xml:"uv"`
	RelativeHumidity struct {
		Text  string `xml:",chardata"`
		Units string `xml:"units,attr"`
	} `xml:"relativeHumidity"`
}

type HourlyForecast struct {
	Text        string `xml:",chardata"`
	DateTimeUTC string `xml:"dateTimeUTC,attr"`
	Condition   string `xml:"condition"`
	IconCode    struct {
		Text   string `xml:",chardata"`
		Format string `xml:"format,attr"`
	} `xml:"iconCode"`
	Temperature struct {
		Value    float32 `xml:",chardata"`
		UnitType string  `xml:"unitType,attr"`
		Units    string  `xml:"units,attr"`
	} `xml:"temperature"`
	Lop struct {
		Text     string `xml:",chardata"`
		Category string `xml:"category,attr"`
		Units    string `xml:"units,attr"`
	} `xml:"lop"`
	WindChill struct {
		Text     string `xml:",chardata"`
		UnitType string `xml:"unitType,attr"`
	} `xml:"windChill"`
	Humidex struct {
		Text     string `xml:",chardata"`
		UnitType string `xml:"unitType,attr"`
	} `xml:"humidex"`
	Wind struct {
		Text  string `xml:",chardata"`
		Speed struct {
			Text     string `xml:",chardata"`
			UnitType string `xml:"unitType,attr"`
			Units    string `xml:"units,attr"`
		} `xml:"speed"`
		Direction struct {
			Text        string `xml:",chardata"`
			WindDirFull string `xml:"windDirFull,attr"`
		} `xml:"direction"`
		Gust struct {
			Text     string `xml:",chardata"`
			UnitType string `xml:"unitType,attr"`
			Units    string `xml:"units,attr"`
		} `xml:"gust"`
	} `xml:"wind"`
}

func (c *canadianWeatherConfig) fetch(url string) (*canadianResponse, error) {
	res, err := http.Get(url)
	if c.debug {
		fmt.Printf("Fetching %s\n", url)
	}
	if err != nil {
		return nil, fmt.Errorf(" unable to get (%s) %v", url, err)
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("unable to read response body (%s): %v", url, err)
	}

	if c.debug {
		fmt.Printf("Response (%s):\n%s\n", url, string(body))
	}

	var resp canadianResponse
	reader := bytes.NewReader(body)
	decoder := xml.NewDecoder(reader)
	decoder.CharsetReader = charset.NewReaderLabel
	err = decoder.Decode(&resp)

	if err != nil {
		return nil, fmt.Errorf("unable to decode the body: %v", err)
	}

	return &resp, nil
}

func (c *canadianWeatherConfig) Setup() {
}

func (c *canadianWeatherConfig) parseDailyForecast(data canadianResponse, numdays int) []iface.Day {
	var forecasts []*iface.Day
	var currDay *iface.Day

	for _, datum := range data.HourlyForecastGroup.HourlyForecast {
		if len(forecasts) >= numdays {
			break
		}

		slot := new(iface.Cond)

		slot.Code = iface.CodeUnknown
		slot.Desc = datum.Condition
		slot.TempC = &datum.Temperature.Value

		year, _ := strconv.ParseInt(datum.DateTimeUTC[0:4], 10, 32)
		month, _ := strconv.ParseInt(datum.DateTimeUTC[4:6], 10, 32)
		day, _ := strconv.ParseInt(datum.DateTimeUTC[6:8], 10, 32)
		hour, _ := strconv.ParseInt(datum.DateTimeUTC[8:10], 10, 32)
		minute, _ := strconv.ParseInt(datum.DateTimeUTC[10:12], 10, 32)

		newTime := time.Date(int(year), time.Month(month), int(day), int(hour), int(minute), 0, 0, time.UTC)

		slot.Time = newTime

		if currDay == nil {
			currDay = new(iface.Day)
			currDay.Date = slot.Time
			forecasts = append(forecasts, currDay)
		}

		if currDay.Date.Day() != slot.Time.Day() {
			currDay = new(iface.Day)
			currDay.Date = slot.Time
			forecasts = append(forecasts, currDay)
		}

		currDay.Slots = append(currDay.Slots, *slot)
	}

	var ret []iface.Day

	for _, day := range forecasts {
		ret = append(ret, *day)
	}

	return ret
}

func (c *canadianWeatherConfig) Fetch(location string, numdays int) iface.Data {
	var ret iface.Data

	resp, err := c.fetch("https://dd.weather.gc.ca/citypage_weather/xml/QC/s0000635_e.xml")
	if err != nil {
		log.Fatalf("Failed to fetch weather data: %v\n", err)
	}

	ret.Current.Code = iface.CodeUnknown
	ret.Current.Desc = resp.CurrentConditions.Condition
	ret.Current.TempC = &resp.CurrentConditions.Temperature.Value

	dateTime := resp.ForecastGroup.DateTime[0]
	ret.Current.Time = time.Date(dateTime.Year, time.Month(dateTime.Month.Value), dateTime.Day.Value, dateTime.Hour, dateTime.Minute, 0, 0, time.UTC)

	ret.Forecast = c.parseDailyForecast(*resp, numdays)

	return ret
}

func init() {
	iface.AllBackends["weather.gc.ca"] = &canadianWeatherConfig{}
}
