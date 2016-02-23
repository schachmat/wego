package frontends

import (
	"encoding/json"
	"flag"
	"log"
	"os"

	"github.com/schachmat/wego/iface"
)

type jsnConfig struct {
	noIndent bool
}

func (c *jsnConfig) Setup() {
	flag.BoolVar(&c.noIndent, "jsn-no-indent", false, "json frontend: do not indent the output")
}

func (c *jsnConfig) Render(r iface.Data, unitSystem iface.UnitSystem) {
	var b []byte
	var err error
	if c.noIndent {
		b, err = json.Marshal(r)
	} else {
		b, err = json.MarshalIndent(r, "", "\t")
	}
	if err != nil {
		log.Fatal(err)
	}
	os.Stdout.Write(b)
}

func init() {
	iface.AllFrontends["json"] = &jsnConfig{}
}
