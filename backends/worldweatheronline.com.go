package backends

import (
	"bytes"
	"encoding/json"
	"flag"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	// needed for some go versions <1.4.2. TODO: Remove this import when golang
	// v1.4.2 or later is in debian stable and the latest Ubuntu LTS release.
	_ "crypto/sha512"

	"github.com/schachmat/wego/iface"
)

type wwoCond struct {
	TmpCor        *int                     `json:"chanceofrain,string"`
	TmpCode       int                      `json:"weatherCode,string"`
	TmpDesc       []struct{ Value string } `json:"weatherDesc"`
	FeelsLikeC    *float32                 `json:",string"`
	PrecipMM      *float32                 `json:"precipMM,string"`
	TmpTempC      *float32                 `json:"tempC,string"`
	TmpTempC2     *float32                 `json:"temp_C,string"`
	TmpTime       *int                     `json:"time,string"`
	VisibleDistKM *float32                 `json:"visibility,string"`
	WindGustKmph  *float32                 `json:",string"`
	WinddirDegree *int                     `json:"winddirDegree,string"`
	WindspeedKmph *float32                 `json:"windspeedKmph,string"`
}

type wwoDay struct {
	Astronomy []struct {
		Moonrise string
		Moonset  string
		Sunrise  string
		Sunset   string
	}
	Date   string
	Hourly []wwoCond
}

type wwoResponse struct {
	Data struct {
		CurCond []wwoCond              `json:"current_condition"`
		Err     []struct{ Msg string } `json:"error"`
		Req     []struct {
			Query string `json:"query"`
			Type  string `json:"type"`
		} `json:"request"`
		Days []wwoDay `json:"weather"`
	} `json:"data"`
}

type wwoCoordinateResp struct {
	Search struct {
		Result []struct {
			Longitude *float32 `json:"longitude,string"`
			Latitude  *float32 `json:"latitude,string"`
		} `json:"result"`
	} `json:"search_api"`
}

type wwoConfig struct {
	apiKey   string
	language string
	debug    bool
}

const (
	wwoSuri = "https://api.worldweatheronline.com/free/v2/search.ashx?"
	wwoWuri = "https://api.worldweatheronline.com/free/v2/weather.ashx?"
)

func wwoParseCond(cond wwoCond, date time.Time) (ret iface.Cond) {
	ret.ChanceOfRainPercent = cond.TmpCor

	codemap := map[int]iface.WeatherCode{
		113: iface.CodeSunny,
		116: iface.CodePartlyCloudy,
		119: iface.CodeCloudy,
		122: iface.CodeVeryCloudy,
		143: iface.CodeFog,
		176: iface.CodeLightShowers,
		179: iface.CodeLightSleetShowers,
		182: iface.CodeLightSleet,
		185: iface.CodeLightSleet,
		200: iface.CodeThunderyShowers,
		227: iface.CodeLightSnow,
		230: iface.CodeHeavySnow,
		248: iface.CodeFog,
		260: iface.CodeFog,
		263: iface.CodeLightShowers,
		266: iface.CodeLightRain,
		281: iface.CodeLightSleet,
		284: iface.CodeLightSleet,
		293: iface.CodeLightRain,
		296: iface.CodeLightRain,
		299: iface.CodeHeavyShowers,
		302: iface.CodeHeavyRain,
		305: iface.CodeHeavyShowers,
		308: iface.CodeHeavyRain,
		311: iface.CodeLightSleet,
		314: iface.CodeLightSleet,
		317: iface.CodeLightSleet,
		320: iface.CodeLightSnow,
		323: iface.CodeLightSnowShowers,
		326: iface.CodeLightSnowShowers,
		329: iface.CodeHeavySnow,
		332: iface.CodeHeavySnow,
		335: iface.CodeHeavySnowShowers,
		338: iface.CodeHeavySnow,
		350: iface.CodeLightSleet,
		353: iface.CodeLightShowers,
		356: iface.CodeHeavyShowers,
		359: iface.CodeHeavyRain,
		362: iface.CodeLightSleetShowers,
		365: iface.CodeLightSleetShowers,
		368: iface.CodeLightSnowShowers,
		371: iface.CodeHeavySnowShowers,
		374: iface.CodeLightSleetShowers,
		377: iface.CodeLightSleet,
		386: iface.CodeThunderyShowers,
		389: iface.CodeThunderyHeavyRain,
		392: iface.CodeThunderySnowShowers,
		395: iface.CodeHeavySnowShowers,
	}
	ret.Code = iface.CodeUnknown
	if val, ok := codemap[cond.TmpCode]; ok {
		ret.Code = val
	}

	if cond.TmpDesc != nil && len(cond.TmpDesc) > 0 {
		ret.Desc = cond.TmpDesc[0].Value
	}

	ret.TempC = cond.TmpTempC2
	if cond.TmpTempC != nil {
		ret.TempC = cond.TmpTempC
	}
	ret.FeelsLikeC = cond.FeelsLikeC

	if cond.PrecipMM != nil {
		p := *cond.PrecipMM / 1000
		ret.PrecipM = &p
	}

	ret.Time = date
	if cond.TmpTime != nil {
		year, month, day := date.Date()
		hour, min := *cond.TmpTime/100, *cond.TmpTime%100
		ret.Time = time.Date(year, month, day, hour, min, 0, 0, time.UTC)
	}

	if cond.VisibleDistKM != nil {
		p := *cond.VisibleDistKM * 1000
		ret.VisibleDistM = &p
	}

	if cond.WinddirDegree != nil && *cond.WinddirDegree >= 0 {
		p := *cond.WinddirDegree % 360
		ret.WinddirDegree = &p
	}

	ret.WindspeedKmph = cond.WindspeedKmph
	ret.WindGustKmph = cond.WindGustKmph

	return
}

func wwoParseDay(day wwoDay, index int) (ret iface.Day) {
	//TODO: Astronomy

	ret.Date = time.Now().Add(time.Hour * 24 * time.Duration(index))
	date, err := time.Parse("2006-01-02", day.Date)
	if err == nil {
		ret.Date = date
	}

	if day.Hourly != nil && len(day.Hourly) > 0 {
		for _, slot := range day.Hourly {
			ret.Slots = append(ret.Slots, wwoParseCond(slot, date))
		}
	}

	return
}

func wwoUnmarshalLang(body []byte, r *wwoResponse, lang string) error {
	var rv map[string]interface{}
	if err := json.Unmarshal(body, &rv); err != nil {
		return err
	}
	if data, ok := rv["data"].(map[string]interface{}); ok {
		if ccs, ok := data["current_condition"].([]interface{}); ok {
			for _, cci := range ccs {
				cc, ok := cci.(map[string]interface{})
				if !ok {
					continue
				}
				langs, ok := cc["lang_"+lang].([]interface{})
				if !ok || len(langs) == 0 {
					continue
				}
				weatherDesc, ok := cc["weatherDesc"].([]interface{})
				if !ok || len(weatherDesc) == 0 {
					continue
				}
				weatherDesc[0] = langs[0]
			}
		}
		if ws, ok := data["weather"].([]interface{}); ok {
			for _, wi := range ws {
				w, ok := wi.(map[string]interface{})
				if !ok {
					continue
				}
				if hs, ok := w["hourly"].([]interface{}); ok {
					for _, hi := range hs {
						h, ok := hi.(map[string]interface{})
						if !ok {
							continue
						}
						langs, ok := h["lang_"+lang].([]interface{})
						if !ok || len(langs) == 0 {
							continue
						}
						weatherDesc, ok := h["weatherDesc"].([]interface{})
						if !ok || len(weatherDesc) == 0 {
							continue
						}
						weatherDesc[0] = langs[0]
					}
				}
			}
		}
	}
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(rv); err != nil {
		return err
	}
	return json.NewDecoder(&buf).Decode(r)
}

func (c *wwoConfig) Setup() {
	flag.StringVar(&c.apiKey, "wwo-api-key", "", "worldweatheronline backend: the api `KEY` to use")
	flag.StringVar(&c.language, "wwo-lang", "en", "worldweatheronline backend: the `LANGUAGE` to request from worldweatheronline")
	flag.BoolVar(&c.debug, "wwo-debug", false, "worldweatheronline backend: print raw requests and responses")
}

func (c *wwoConfig) getCoordinatesFromAPI(queryParams []string, res chan *iface.LatLon) {
	var coordResp wwoCoordinateResp
	requri := wwoSuri + strings.Join(queryParams, "&")
	hres, err := http.Get(requri)
	if err != nil {
		log.Println("Unable to fetch geo location:", err)
		res <- nil
		return
	} else if hres.StatusCode != 200 {
		log.Println("Unable to fetch geo location: http status", hres.StatusCode)
		res <- nil
		return
	}
	defer hres.Body.Close()

	body, err := io.ReadAll(hres.Body)
	if err != nil {
		log.Println("Unable to read geo location data:", err)
		res <- nil
		return
	}

	if c.debug {
		log.Println("Geo location request:", requri)
		log.Println("Geo location response:", string(body))
	}

	if err = json.Unmarshal(body, &coordResp); err != nil {
		log.Println("Unable to unmarshal geo location data:", err)
		res <- nil
		return
	}

	r := coordResp.Search.Result
	if len(r) < 1 || r[0].Latitude == nil || r[0].Longitude == nil {
		log.Println("Malformed geo location response")
		res <- nil
		return
	}

	res <- &iface.LatLon{Latitude: *r[0].Latitude, Longitude: *r[0].Longitude}
}

func (c *wwoConfig) Fetch(loc string, numdays int) iface.Data {
	var params []string
	var resp wwoResponse
	var ret iface.Data
	coordChan := make(chan *iface.LatLon)

	if len(c.apiKey) == 0 {
		log.Fatal("No API key specified. Setup instructions are in the README.")
	}
	params = append(params, "key="+c.apiKey)

	if len(loc) > 0 {
		params = append(params, "q="+url.QueryEscape(loc))
	}
	params = append(params, "format=json")
	params = append(params, "num_of_days="+strconv.Itoa(numdays))
	params = append(params, "tp=3")

	go c.getCoordinatesFromAPI(params, coordChan)

	if c.language != "" {
		params = append(params, "lang="+c.language)
	}
	requri := wwoWuri + strings.Join(params, "&")

	res, err := http.Get(requri)
	if err != nil {
		log.Fatal("Unable to get weather data: ", err)
	} else if res.StatusCode != 200 {
		log.Fatal("Unable to get weather data: http status ", res.StatusCode)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	if c.debug {
		log.Println("Weather request:", requri)
		log.Println("Weather response:", string(body))
	}

	if c.language == "" {
		if err = json.Unmarshal(body, &resp); err != nil {
			log.Println(err)
		}
	} else {
		if err = wwoUnmarshalLang(body, &resp, c.language); err != nil {
			log.Println(err)
		}
	}

	if resp.Data.Req == nil || len(resp.Data.Req) < 1 {
		if resp.Data.Err != nil && len(resp.Data.Err) >= 1 {
			log.Fatal(resp.Data.Err[0].Msg)
		}
		log.Fatal("Malformed response.")
	}

	ret.Location = resp.Data.Req[0].Type + ": " + resp.Data.Req[0].Query
	ret.GeoLoc = <-coordChan

	if resp.Data.CurCond != nil && len(resp.Data.CurCond) > 0 {
		ret.Current = wwoParseCond(resp.Data.CurCond[0], time.Now())
	}

	if resp.Data.Days != nil && numdays > 0 {
		for i, day := range resp.Data.Days {
			ret.Forecast = append(ret.Forecast, wwoParseDay(day, i))
		}
	}

	return ret
}

func init() {
	iface.AllBackends["worldweatheronline"] = &wwoConfig{}
}
