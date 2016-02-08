package backends

import (
	"github.com/schachmat/wego/iface"
)

type Backend interface {
	Setup()
	Fetch(string, int) iface.Resp
}

var (
	All = make(map[string]Backend)
)
