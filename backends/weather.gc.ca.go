package backends

import (
	"github.com/schachmat/wego/iface"
)

type canadianWeatherConfig struct {
}

func (c *canadianWeatherConfig) Setup() {
}

func (c *canadianWeatherConfig) Fetch(location string, numdays int) iface.Data {
	var ret iface.Data
	return ret
}

func init() {
	iface.AllBackends["canadianweather"] = &canadianWeatherConfig{}
}
