package backends

import (
	"github.com/schachmat/wego/iface"
)

type Backend struct {
	Setup func(map[string]interface{})
	Fetch func(map[string]interface{}, string, int) iface.Resp
}

var (
	All = make(map[string]Backend)
)
