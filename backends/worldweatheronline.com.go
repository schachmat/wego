package backends

import (
	"bytes"
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/schachmat/wego/iface"
)

type worldweatheronline struct {
	wwoApiKey string
	wwoLanguage string
}

const (
	wuri = "https://api.worldweatheronline.com/free/v2/weather.ashx?"
	suri = "https://api.worldweatheronline.com/free/v2/search.ashx?"
)

func unmarshalLang(body []byte, r *iface.Resp, lang string) error {
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

func (c *worldweatheronline) Setup() {
	flag.StringVar(&c.wwoApiKey, "wwo-api-key", "", "wwo backend: the api `KEY` to use")
	flag.StringVar(&c.wwoLanguage, "wwo-lang", "en", "wwo backend: the `LANGUAGE` to request from wwo")
}

func (c *worldweatheronline) Fetch(loc string, numdays int) (ret iface.Resp) {
	var params []string

	if len(c.wwoApiKey) == 0 {
		log.Fatal("No API key specified. Setup instructions are in the README.")
	}
	params = append(params, "key="+c.wwoApiKey)

	if len(loc) > 0 {
		params = append(params, "q="+url.QueryEscape(loc))
	}
	params = append(params, "format=json")
	params = append(params, "num_of_days="+strconv.Itoa(numdays))
	params = append(params, "tp=3")
	if c.wwoLanguage != "" {
		params = append(params, "lang="+c.wwoLanguage)
	}

	res, err := http.Get(wuri + strings.Join(params, "&"))
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	if c.wwoLanguage == "" {
		if err = json.Unmarshal(body, &ret); err != nil {
			log.Println(err)
		}
	} else {
		if err = unmarshalLang(body, &ret, c.wwoLanguage); err != nil {
			log.Println(err)
		}
	}

	if ret.Data.Req == nil || len(ret.Data.Req) < 1 {
		if ret.Data.Err != nil && len(ret.Data.Err) >= 1 {
			log.Fatal(ret.Data.Err[0].Msg)
		}
		log.Fatal("Malformed response.")
	}

	if numdays == 0 {
		ret.Data.Weather = []iface.Weather{}
	}
	return
}

func init() {
	All["worldweatheronline.com"] = &worldweatheronline{}
}
