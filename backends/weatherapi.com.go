package backends

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/schachmat/wego/iface"
)

type weatherApiResponse struct {
	Location struct {
		Name    string `json:"name"`
		Country string `json:"country"`
	} `json:"location"`
	Forcast struct {
		List []forcastBlock `json:"forcastday"`
	} `json:"forcast"`
}

type forcastBlock struct {
	Date time.Time `json:"date"`
	Day  struct {
		TempC        float32 `json:"avgtemp_c"`
		Humidity     int     `json:"avghumidity"`
		MaxWindSpeed float32 `json:"maxwind_kph"`
		Weather      struct {
			Description string `json:"text"`
			Code        int    `json:"code"`
		} `json:"condition"`
	} `json:"day"`
}

type weatherApiConfig struct {
	apiKey string
	debug  bool
}

func (c *weatherApiConfig) Setup() {
	flag.StringVar(&c.apiKey, "wth-api-key", "", "weatherapi backend: the api `Key` to use")
	flag.BoolVar(&c.debug, "wth-debug", false, "weatherapi backend: print raw requests and responses")
}

func (c *weatherApiConfig) fetch(url string) (*weatherApiResponse, error) {
	res, err := http.Get(url)
	if c.debug {
		fmt.Printf("Fetching %s\n", url)
	}
	if err != nil {
		return nil, fmt.Errorf("Unable to get (%s) %v", url, err)
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("Unable to read response body (%s): %v", url, err)
	}

	if c.debug {
		fmt.Printf("Response (%s):\n%s\n", url, string(body))
	}

	var resp weatherApiResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("Unable to unmarshal response (%s): %v\nThe json body is: %s", url, err, string(body))
	}

	return &resp, nil
}

func (c *weatherApiConfig) parseCond(forcastInfo forcastBlock) (iface.Cond, error) {
	var ret iface.Cond
	codemap := map[int]iface.WeatherCode{
		1000: iface.CodeSunny,
		1003: iface.CodePartlyCloudy,
		1006: iface.CodeCloudy,
		1009: iface.CodeVeryCloudy,
		1030: iface.CodeVeryCloudy,
		1063: iface.CodeLightRain,
		1066: iface.CodeLightSnowShowers,
		1069: iface.CodeLightSnowShowers,
		1071: iface.CodeLightShowers,
		1087: iface.CodeThunderyShowers,
		1114: iface.CodeHeavySnow,
		1117: iface.CodeHeavySnow,
		1135: iface.CodeFog,
		1147: iface.CodeFog,
		1150: iface.CodeLightRain,
		1153: iface.CodeLightRain,
		1168: iface.CodeLightRain,
		1171: iface.CodeHeavyRain,
		1180: iface.CodeLightRain,
		1183: iface.CodeLightRain,
		1186: iface.CodeHeavyRain,
		1189: iface.CodeHeavyRain,
		1192: iface.CodeHeavyRain,
		1195: iface.CodeHeavyRain,
		1198: iface.CodeLightRain,
		1201: iface.CodeHeavyRain,
		1204: iface.CodeLightSleet,
		1207: iface.CodeLightSleetShowers,
		1210: iface.CodeLightSnow,
		1213: iface.CodeLightSnow,
		1216: iface.CodeHeavySnow,
		1219: iface.CodeHeavySnow,
		1222: iface.CodeHeavySnow,
		1225: iface.CodeHeavySnow,
		1237: iface.CodeHeavySnow,
		1240: iface.CodeLightShowers,
		1243: iface.CodeHeavyShowers,
		1246: iface.CodeThunderyShowers,
		1249: iface.CodeLightSleetShowers,
		1252: iface.CodeLightSleetShowers,
		1255: iface.CodeLightSnowShowers,
		1258: iface.CodeHeavySnowShowers,
		1261: iface.CodeLightSnowShowers,
		1264: iface.CodeHeavySnowShowers,
		1273: iface.CodeThunderyShowers,
		1276: iface.CodeThunderyHeavyRain,
		1279: iface.CodeThunderySnowShowers,
		1282: iface.CodeThunderySnowShowers,
	}

	ret.Code = iface.CodeUnknown
	ret.Desc = forcastInfo.Day.Weather.Description
	ret.Humidity = &(forcastInfo.Day.Humidity)
	ret.TempC = &(forcastInfo.Day.TempC)
	ret.WindspeedKmph = &(forcastInfo.Day.MaxWindSpeed)

	if val, ok := codemap[forcastInfo.Day.Weather.Code]; ok {
		ret.Code = val
	}

	return ret, nil
}

func (c *weatherApiConfig) Fetch(location string, numdays int) iface.Data {

	return iface.Data{}
}

func init() {
	iface.AllBackends["weatherapi"] = &weatherApiConfig{}
}
