package main

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestParseAliases(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected map[string]string
	}{
		{
			name:     "empty string",
			input:    "",
			expected: map[string]string{},
		},
		{
			name:     "whitespace only",
			input:    "   \t\n  ",
			expected: map[string]string{},
		},
		{
			name:     "single alias with equal sign",
			input:    "home=27.35,-82.49",
			expected: map[string]string{"home": "27.35,-82.49"},
		},
		{
			name:     "single alias with colon",
			input:    "home:27.35,-82.49",
			expected: map[string]string{"home": "27.35,-82.49"},
		},
		{
			name:  "multiple semicolon-separated aliases",
			input: "home=27.35,-82.49;work=28.4311,-81.3077",
			expected: map[string]string{
				"home": "27.35,-82.49",
				"work": "28.4311,-81.3077",
			},
		},
		{
			name:  "multiple comma-separated aliases with coordinate commas",
			input: "home=27.35,-82.49,work=28.4311,-81.3077",
			expected: map[string]string{
				"home": "27.35,-82.49",
				"work": "28.4311,-81.3077",
			},
		},
		{
			name:  "semicolon-separated with spaces and city names",
			input: "  home = New York, NY ;  work = London, UK  ",
			expected: map[string]string{
				"home": "New York, NY",
				"work": "London, UK",
			},
		},
		{
			name:  "multiline input with comments and blank lines",
			input: "# User location aliases\nhome=27.35,-82.49\n\n# Vacation spot\nbeach=25.76,-80.19\n",
			expected: map[string]string{
				"home":  "27.35,-82.49",
				"beach": "25.76,-80.19",
			},
		},
		{
			name:  "case normalization of alias keys",
			input: "HOME=27.35,-82.49;WorkPlace=London",
			expected: map[string]string{
				"home":      "27.35,-82.49",
				"workplace": "London",
			},
		},
		{
			name:     "invalid entries skipped",
			input:    "noequal;=novalue;nokey=;valid=12.34,56.78",
			expected: map[string]string{"valid": "12.34,56.78"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := parseAliases(tc.input)
			if len(got) == 0 && len(tc.expected) == 0 {
				return
			}
			if !reflect.DeepEqual(got, tc.expected) {
				t.Errorf("parseAliases(%q) = %v, expected %v", tc.input, got, tc.expected)
			}
		})
	}
}

func TestLookupAlias(t *testing.T) {
	aliases := map[string]string{
		"home":      "27.35,-82.49",
		"work":      "28.4311,-81.3077",
		"office":    "work",
		"headquar":  "office",
		"loopa":     "loopb",
		"loopb":     "loopa",
		"airport":   "kmco",
		"kmco":      "28.4311,-81.3077",
	}

	tests := []struct {
		name        string
		aliases     map[string]string
		query       string
		expectedVal string
		expectedOk  bool
	}{
		{
			name:        "nil aliases map",
			aliases:     nil,
			query:       "home",
			expectedVal: "",
			expectedOk:  false,
		},
		{
			name:        "empty query",
			aliases:     aliases,
			query:       "",
			expectedVal: "",
			expectedOk:  false,
		},
		{
			name:        "exact match",
			aliases:     aliases,
			query:       "home",
			expectedVal: "27.35,-82.49",
			expectedOk:  true,
		},
		{
			name:        "case-insensitive upper case",
			aliases:     aliases,
			query:       "HOME",
			expectedVal: "27.35,-82.49",
			expectedOk:  true,
		},
		{
			name:        "case-insensitive mixed case",
			aliases:     aliases,
			query:       "HoMe",
			expectedVal: "27.35,-82.49",
			expectedOk:  true,
		},
		{
			name:        "query with surrounding whitespace",
			aliases:     aliases,
			query:       "  home  ",
			expectedVal: "27.35,-82.49",
			expectedOk:  true,
		},
		{
			name:        "non-existent alias",
			aliases:     aliases,
			query:       "unknown",
			expectedVal: "",
			expectedOk:  false,
		},
		{
			name:        "alias chain resolution",
			aliases:     aliases,
			query:       "airport",
			expectedVal: "28.4311,-81.3077",
			expectedOk:  true,
		},
		{
			name:        "multi-hop alias chain resolution",
			aliases:     aliases,
			query:       "headquar",
			expectedVal: "28.4311,-81.3077",
			expectedOk:  true,
		},
		{
			name:        "cyclic alias resolution does not panic or loop",
			aliases:     aliases,
			query:       "loopa",
			expectedVal: "loopb",
			expectedOk:  true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			val, ok := lookupAlias(tc.query, tc.aliases)
			if ok != tc.expectedOk || val != tc.expectedVal {
				t.Errorf("lookupAlias(%q) = (%q, %v), expected (%q, %v)",
					tc.query, val, ok, tc.expectedVal, tc.expectedOk)
			}
		})
	}
}

func TestResolveAlias(t *testing.T) {
	aliases := map[string]string{
		"home": "27.35,-82.49",
		"work": "London",
	}

	tests := []struct {
		name     string
		aliases  map[string]string
		query    string
		expected string
	}{
		{
			name:     "resolves matching alias",
			aliases:  aliases,
			query:    "home",
			expected: "27.35,-82.49",
		},
		{
			name:     "resolves case-insensitively",
			aliases:  aliases,
			query:    "HOME",
			expected: "27.35,-82.49",
		},
		{
			name:     "fallback for unknown alias returns input",
			aliases:  aliases,
			query:    "Paris",
			expected: "Paris",
		},
		{
			name:     "fallback for raw coordinates returns coordinates",
			aliases:  aliases,
			query:    "40.748,-73.985",
			expected: "40.748,-73.985",
		},
		{
			name:     "fallback with nil aliases map returns input",
			aliases:  nil,
			query:    "London",
			expected: "London",
		},
		{
			name:     "fallback for empty input returns empty",
			aliases:  aliases,
			query:    "",
			expected: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := resolveAlias(tc.query, tc.aliases)
			if got != tc.expected {
				t.Errorf("resolveAlias(%q) = %q, expected %q", tc.query, got, tc.expected)
			}
		})
	}
}

func TestResolveLocation(t *testing.T) {
	aliases := map[string]string{
		"home": "27.35,-82.49",
		"work": "28.4311,-81.3077",
	}

	tests := []struct {
		name     string
		loc      string
		alias    string
		aliases  map[string]string
		expected string
	}{
		{
			name:     "alias flag provided and resolves",
			loc:      "40.748,-73.985",
			alias:    "home",
			aliases:  aliases,
			expected: "27.35,-82.49",
		},
		{
			name:     "alias flag provided with uppercase resolves",
			loc:      "40.748,-73.985",
			alias:    "HOME",
			aliases:  aliases,
			expected: "27.35,-82.49",
		},
		{
			name:     "alias flag provided but unknown falls back to alias value",
			loc:      "40.748,-73.985",
			alias:    "unknown-place",
			aliases:  aliases,
			expected: "unknown-place",
		},
		{
			name:     "positional location matches alias and resolves",
			loc:      "home",
			alias:    "",
			aliases:  aliases,
			expected: "27.35,-82.49",
		},
		{
			name:     "positional location with uppercase matches alias and resolves",
			loc:      "HOME",
			alias:    "",
			aliases:  aliases,
			expected: "27.35,-82.49",
		},
		{
			name:     "positional location not matching alias falls back to location",
			loc:      "London",
			alias:    "",
			aliases:  aliases,
			expected: "London",
		},
		{
			name:     "coordinates as location fall back to coordinates",
			loc:      "40.748,-73.985",
			alias:    "",
			aliases:  aliases,
			expected: "40.748,-73.985",
		},
		{
			name:     "empty alias and empty location fall back to empty",
			loc:      "",
			alias:    "",
			aliases:  aliases,
			expected: "",
		},
		{
			name:     "alias flag takes precedence over positional location",
			loc:      "London",
			alias:    "work",
			aliases:  aliases,
			expected: "28.4311,-81.3077",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := resolveLocation(tc.loc, tc.alias, tc.aliases)
			if got != tc.expected {
				t.Errorf("resolveLocation(%q, %q) = %q, expected %q", tc.loc, tc.alias, got, tc.expected)
			}
		})
	}
}

func TestLoadAliasesFile(t *testing.T) {
	t.Run("nonexistent file returns empty map without error", func(t *testing.T) {
		got := loadAliasesFile("/path/to/nonexistent/file")
		if len(got) != 0 {
			t.Errorf("loadAliasesFile for nonexistent file returned non-empty map: %v", got)
		}
	})

	t.Run("valid alias file loaded correctly", func(t *testing.T) {
		tmpDir := t.TempDir()
		aliasFilePath := filepath.Join(tmpDir, "aliases")
		content := "# Configured aliases\nhome=27.35,-82.49\n\nWORK=28.4311,-81.3077\n"
		if err := os.WriteFile(aliasFilePath, []byte(content), 0600); err != nil {
			t.Fatalf("failed to write test alias file: %v", err)
		}

		got := loadAliasesFile(aliasFilePath)
		expected := map[string]string{
			"home": "27.35,-82.49",
			"work": "28.4311,-81.3077",
		}
		if !reflect.DeepEqual(got, expected) {
			t.Errorf("loadAliasesFile(%s) = %v, expected %v", aliasFilePath, got, expected)
		}
	})
}
