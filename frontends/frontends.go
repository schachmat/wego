package frontends

import (
	"github.com/schachmat/wego/iface"
)

type Frontend interface {
	Setup()
	Render(iface.Resp)
}

var (
	All = make(map[string]Frontend)
)
