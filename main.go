package main

import (
	"crypto/sha256"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/muesli/mango"
	"github.com/muesli/mango/mflag"
	"github.com/muesli/roff"
	"github.com/schachmat/ingo"
	_ "github.com/schachmat/wego/backends"
	_ "github.com/schachmat/wego/frontends"
	"github.com/schachmat/wego/iface"
)

type cacheEntry struct {
	Expires time.Time  `json:"expires"`
	Data    iface.Data `json:"data"`
}

func cacheFilePath(backend, location string, numdays int) string {
	key := fmt.Sprintf("%s|%s|%d", backend, location, numdays)
	hash := sha256.Sum256([]byte(key))
	return filepath.Join(os.TempDir(), fmt.Sprintf("wego-cache-%x.json", hash))
}

func loadCache(path string) (iface.Data, bool) {
	f, err := os.Open(path)
	if err != nil {
		return iface.Data{}, false
	}
	defer func() { _ = f.Close() }()

	var entry cacheEntry
	if err := json.NewDecoder(f).Decode(&entry); err != nil {
		return iface.Data{}, false
	}
	if time.Now().After(entry.Expires) {
		return iface.Data{}, false
	}
	return entry.Data, true
}

func saveCache(path string, data iface.Data, ttl time.Duration) {
	entry := cacheEntry{
		Expires: time.Now().Add(ttl),
		Data:    data,
	}
	f, err := os.Create(path)
	if err != nil {
		log.Printf("Warning: could not write cache file %s: %v", path, err)
		return
	}
	encErr := json.NewEncoder(f).Encode(entry)
	closeErr := f.Close()
	if encErr != nil || closeErr != nil {
		if encErr != nil {
			log.Printf("Warning: could not encode cache data to %s: %v", path, encErr)
		} else {
			log.Printf("Warning: could not close cache file %s: %v", path, closeErr)
		}
		if removeErr := os.Remove(path); removeErr != nil {
			log.Printf("Warning: could not remove corrupt cache file %s: %v", path, removeErr)
		}
	}
}

func pluginLists() {
	bEnds := make([]string, 0, len(iface.AllBackends))
	for name := range iface.AllBackends {
		bEnds = append(bEnds, name)
	}
	sort.Strings(bEnds)

	fEnds := make([]string, 0, len(iface.AllFrontends))
	for name := range iface.AllFrontends {
		fEnds = append(fEnds, name)
	}
	sort.Strings(fEnds)

	fmt.Fprintln(os.Stderr, "Available backends:", strings.Join(bEnds, ", "))
	fmt.Fprintln(os.Stderr, "Available frontends:", strings.Join(fEnds, ", "))
}

func main() {
	// initialize backends and frontends (flags and default config)
	for _, be := range iface.AllBackends {
		be.Setup()
	}
	for _, fe := range iface.AllFrontends {
		fe.Setup()
	}

	// initialize global flags and default config
	location := flag.String("location", "40.748,-73.985", "`LOCATION` to be queried")
	flag.StringVar(location, "l", "40.748,-73.985", "`LOCATION` to be queried (shorthand)")
	numdays := flag.Int("days", 3, "`NUMBER` of days of weather forecast to be displayed")
	flag.IntVar(numdays, "d", 3, "`NUMBER` of days of weather forecast to be displayed (shorthand)")
	unitSystem := flag.String("units", "metric", "`UNITSYSTEM` to use for output.\n    \tChoices are: metric, imperial, si, metric-ms")
	flag.StringVar(unitSystem, "u", "metric", "`UNITSYSTEM` to use for output. (shorthand)\n    \tChoices are: metric, imperial, si, metric-ms")
	selectedBackend := flag.String("backend", "openweathermap", "`BACKEND` to be used")
	flag.StringVar(selectedBackend, "b", "openweathermap", "`BACKEND` to be used (shorthand)")
	selectedFrontend := flag.String("frontend", "ascii-art-table", "`FRONTEND` to be used")
	flag.StringVar(selectedFrontend, "f", "ascii-art-table", "`FRONTEND` to be used (shorthand)")
	flag.Bool("man", false, "Generate man page and print to stdout")
	cacheTTL := flag.Duration("cache-ttl", time.Hour, "`DURATION` to cache weather data on disk (0 to disable)")

	// print out a list of all backends and frontends in the usage
	tmpUsage := flag.Usage
	flag.Usage = func() {
		tmpUsage()
		pluginLists()
	}

	// generate and print man page if requested, before config parsing so that
	// a missing or malformed config file does not prevent the man page from showing
	for _, arg := range os.Args[1:] {
		if arg == "-man" || arg == "--man" {
			manPage := mango.NewManPage(1, "wego", "display the weather in your terminal").
				WithLongDescription("wego is a weather client for the terminal that shows "+
					"the current and forecasted weather conditions using various backends.\n"+
					"Configuration is stored in a config file (default: ~/.wegorc) and can "+
					"also be provided via command-line flags.\n"+
					"A backend API key is required for most backends.").
				WithSection("Configuration", "wego stores its configuration in ~/.wegorc by default. "+
					"The config file is created on the first run with default values. "+
					"Each flag listed below corresponds to a config file key. "+
					"Command-line flags take precedence over config file values.").
				WithSection("Copyright", "(C) The wego contributors.\nReleased under ISC license.")
			flag.VisitAll(mflag.FlagVisitor(manPage))
			fmt.Println(manPage.Build(roff.NewDocument()))
			os.Exit(0)
		}
	}

	// read/write config and parse flags
	if err := ingo.Parse("wego"); err != nil {
		log.Fatalf("Error parsing config: %v", err)
	}

	// non-flag shortcut arguments overwrite possible flag arguments
	for _, arg := range flag.Args() {
		if v, err := strconv.Atoi(arg); err == nil && len(arg) == 1 {
			*numdays = v
		} else {
			*location = arg
		}
	}

	// get selected backend and fetch the weather data from it
	be, ok := iface.AllBackends[*selectedBackend]
	if !ok {
		log.Fatalf("Could not find selected backend \"%s\"", *selectedBackend)
	}

	var r iface.Data
	cachePath := cacheFilePath(*selectedBackend, *location, *numdays)
	if *cacheTTL > 0 {
		if cached, hit := loadCache(cachePath); hit {
			r = cached
		} else {
			r = be.Fetch(*location, *numdays)
			saveCache(cachePath, r, *cacheTTL)
		}
	} else {
		r = be.Fetch(*location, *numdays)
	}

	// set unit system
	unit := iface.UnitsMetric
	if *unitSystem == "imperial" {
		unit = iface.UnitsImperial
	} else if *unitSystem == "si" {
		unit = iface.UnitsSi
	} else if *unitSystem == "metric-ms" {
		unit = iface.UnitsMetricMs
	}

	// get selected frontend and render the weather data with it
	fe, ok := iface.AllFrontends[*selectedFrontend]
	if !ok {
		log.Fatalf("Could not find selected frontend \"%s\"", *selectedFrontend)
	}
	fe.Render(r, unit)
}
