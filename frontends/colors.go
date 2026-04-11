package frontends

import (
	"flag"
	"os"
	"strconv"
	"strings"
)

var monochrome bool

func init() {
	flag.BoolVar(&monochrome, "monochrome", false, "Monochrome output")
}

// isMonochrome returns true if color output should be disabled, either because
// the --monochrome flag was set or because the NO_COLOR environment variable
// (https://no-color.org/) is present.
func isMonochrome() bool {
	_, noColorSet := os.LookupEnv("NO_COLOR")
	return monochrome || noColorSet
}

// darkBackground returns true if the terminal likely has a dark background.
// It checks the COLORFGBG environment variable which is set by some terminals
// (e.g. rxvt, xterm). The variable's format is "foreground;background" or
// "foreground;default;background" where the color numbers match the 16-color
// palette (0=black … 8=dark gray, 9-15=bright colors). Background colors 0-8
// indicate a dark background, 9-15 indicate a light background.
// Defaults to true (dark background) when the variable is absent or
// unrecognizable, since dark terminals are the most common use case.
func darkBackground() bool {
	colorfgbg := os.Getenv("COLORFGBG")
	if colorfgbg == "" {
		return true
	}
	parts := strings.Split(colorfgbg, ";")
	bg := parts[len(parts)-1]
	if bg == "default" {
		return true
	}
	bgNum, err := strconv.Atoi(bg)
	if err != nil {
		return true
	}
	// In the 16-color palette: 0-6 are dark colors (black through dark cyan),
	// 7 is white/light-gray (treated as light), 8 is dark gray (treated as dark),
	// and 9-15 are bright/light variants (treated as light).
	return bgNum < 7 || bgNum == 8
}
