package main

import (
	"flag"
	"log"
	"os"
	"os/user"
	"path"
	"strconv"

	"github.com/schachmat/ingo"
	"github.com/schachmat/wego/backends"
	"github.com/schachmat/wego/frontends"
)

func main() {
	configpath := os.Getenv("WEGORC")
	if configpath == "" {
		usr, err := user.Current()
		if err != nil {
			log.Fatalf("%v\nYou can set the environment variable WEGORC to point to your config file as a workaround.", err)
		}
		configpath = path.Join(usr.HomeDir, ".wegorc")
	}

	// initialize backends and frontends (flags and default config)
	for _, be := range backends.All {
		be.Setup()
	}
	for _, fe := range frontends.All {
		fe.Setup()
	}

	// initialize global flags and default config
	numdays := flag.Int("days", 3, "`NUMBER` of days of weather forecast to be displayed")
	location := flag.String("city", "New York", "`LOCATION` to be queried")
	selectedBackend := flag.String("backend", "worldweatheronline.com", "`BACKEND` to be used")
	selectedFrontend := flag.String("frontend", "ascii-art-table", "`FRONTEND` to be used")

	// read/write config and parse flags
	ingo.Parse(configpath)

	// non-flag shortcut arguments overwrite possible flag arguments
	for _, arg := range flag.Args() {
		if v, err := strconv.Atoi(arg); err == nil && len(arg) == 1 {
			*numdays = v
		} else {
			*location = arg
		}
	}

	// get selected backend and fetch the weather data from it
	be, ok := backends.All[*selectedBackend]
	if !ok {
		log.Fatalf("Could not find selected backend \"%s\"", *selectedBackend)
	}
	r := be.Fetch(*location, *numdays)

	// get selected frontend and render the weather data with it
	fe, ok := frontends.All[*selectedFrontend]
	if !ok {
		log.Fatalf("Could not find selected frontend \"%s\"", *selectedFrontend)
	}
	fe.Render(r)
}
