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
	} else {
		if err = json.NewDecoder(&buf).Decode(r); err != nil {
			return err
		}
	}
	return nil
}

func wwoSetup(conf map[string]interface{}) {
	conf["APIKey"] = flag.String("wwo-api-key", "", "wwo backend: API key")
	conf["Lang"] = flag.String("wwo-lang", "en", "wwo backend: language")
}

func wwoFlags(conf map[string]interface{}) {
	conf["APIKey"] = flag.String("wwo-api-key", "", "wwo backend: API key")
	conf["Lang"] = flag.String("wwo-lang", "en", "wwo backend: language")
}

func wwoFetch(conf map[string]interface{}, loc string, numdays int) (ret iface.Resp) {
	var params []string

	APIKey, ok := conf["APIKey"].(string)
	if !ok {
		APIKey = *conf["APIKey"].(*string)
	}

	Lang, ok := conf["Lang"].(string)
	if !ok {
		Lang = *conf["Lang"].(*string)
	}

	if len(APIKey) == 0 {
		log.Fatal("No API key specified. Setup instructions are in the README.")
	}
	params = append(params, "key="+APIKey)

	// non-flag shortcut arguments overwrite possible flag arguments
	for _, arg := range flag.Args() {
		if v, err := strconv.Atoi(arg); err == nil && len(arg) == 1 {
			numdays = v
		} else {
			loc = arg
		}
	}

	if len(loc) > 0 {
		params = append(params, "q="+url.QueryEscape(loc))
	}
	params = append(params, "format=json")
	params = append(params, "num_of_days="+strconv.Itoa(numdays))
	params = append(params, "tp=3")
	if Lang != "" {
		params = append(params, "lang="+Lang)
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

	if Lang == "" {
		if err = json.Unmarshal(body, &ret); err != nil {
			log.Println(err)
		}
	} else {
		if err = unmarshalLang(body, &ret, Lang); err != nil {
			log.Println(err)
		}
	}
	return
}

func init() {
	All["worldweatheronline.com"] = Backend{wwoSetup, wwoFetch}
}
