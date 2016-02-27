package main

import (
	"flag"
	"log"
	"strconv"

	"github.com/schachmat/ingo"
	_ "github.com/schachmat/wego/backends"
	_ "github.com/schachmat/wego/frontends"
	"github.com/schachmat/wego/iface"
)

func main() {
	// initialize backends and frontends (flags and default config)
	for _, be := range iface.AllBackends {
		be.Setup()
	}
	for _, fe := range iface.AllFrontends {
		fe.Setup()
	}

	// initialize global flags and default config
	numdays := flag.Int("days", 3, "`NUMBER` of days of weather forecast to be displayed")
	numLocations := flag.Int("locations", 1, "`NUMBER` of locations of weather forecast to be displayed")
	location1 := flag.String("city1", "New York", "`LOCATION1` to be queried")
	location2 := flag.String("city2", "New York", "`LOCATION2` to be queried")
	unitSystem := flag.String("units", "metric", "`UNITSYSTEM` to use for output.\n    \tChoices are: metric, imperial, si")
	selectedBackend := flag.String("backend", "worldweatheronline.com", "`BACKEND` to be used")
	selectedFrontend := flag.String("frontend", "ascii-art-table", "`FRONTEND` to be used")

	// read/write config and parse flags
	if err := ingo.Parse("wego"); err != nil {
		log.Fatalf("Error parsing config: %v", err)
	}

	// non-flag shortcut arguments overwrite possible flag arguments
	for _, arg := range flag.Args() {
		if v, err := strconv.Atoi(arg); err == nil && len(arg) == 1 {
			*numdays = v
			*numLocations = v
		} else {
			*location1 = arg
			*location2 = arg
		}
	}

	// get selected backend and fetch the weather data from it
	be, ok := iface.AllBackends[*selectedBackend]
	if !ok {
		log.Fatalf("Could not find selected backend \"%s\"", *selectedBackend)
	}

	// set unit system
	unit := iface.UnitsMetric
	if *unitSystem == "imperial" {
		unit = iface.UnitsImperial
	} else if *unitSystem == "si" {
		unit = iface.UnitsSi
	}

	// get selected frontend and render the weather data with it
	fe, ok := iface.AllFrontends[*selectedFrontend]
	if !ok {
		log.Fatalf("Could not find selected frontend \"%s\"", *selectedFrontend)
	}

	cities := []*string{location1, location2}
	for i := 0; i < *numLocations; i++ {
		r := be.Fetch(*cities[i], *numdays)
		fe.Render(r, unit)
	}
}
