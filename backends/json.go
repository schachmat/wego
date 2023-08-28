package backends

import (
	"encoding/json"
	"io/ioutil"
	"log"

	"github.com/meatcoder/wedash/iface"
)

type jsnConfig struct {
}

func (c *jsnConfig) Setup() {
}

// Fetch will try to open the file specified in the location string argument and
// read it as json content to fill the data. The numdays argument will only work
// to further limit the amount of days in the output. It obviously cannot
// produce more data than is available in the file.
func (c *jsnConfig) Fetch(loc string, numdays int) (ret iface.Data) {
	b, err := ioutil.ReadFile(loc)
	if err != nil {
		log.Fatal(err)
	}

	err = json.Unmarshal(b, &ret)
	if err != nil {
		log.Fatal(err)
	}

	if len(ret.Forecast) > numdays {
		ret.Forecast = ret.Forecast[:numdays]
	}
	return
}

func init() {
	iface.AllBackends["json"] = &jsnConfig{}
}
