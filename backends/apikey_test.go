package backends

import (
	"flag"
	"os"
	"path/filepath"
	"testing"
)

func TestResolveAPIKey_InlineKey(t *testing.T) {
	key, err := ResolveAPIKey("inline-secret", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if key != "inline-secret" {
		t.Fatalf("expected 'inline-secret', got '%s'", key)
	}
}

func TestResolveAPIKey_KeyFromFile(t *testing.T) {
	tmpDir := t.TempDir()
	keyPath := filepath.Join(tmpDir, "key.txt")
	expectedKey := "test-secret-key-12345"

	if err := os.WriteFile(keyPath, []byte(expectedKey), 0600); err != nil {
		t.Fatalf("failed to create temp key file: %v", err)
	}

	key, err := ResolveAPIKey("", keyPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if key != expectedKey {
		t.Fatalf("expected '%s', got '%s'", expectedKey, key)
	}
}

func TestResolveAPIKey_TrailingWhitespaceAndNewlines(t *testing.T) {
	tmpDir := t.TempDir()

	testCases := []struct {
		name     string
		content  string
		expected string
	}{
		{
			name:     "trailing newline",
			content:  "my-secret-key\n",
			expected: "my-secret-key",
		},
		{
			name:     "trailing CRLF",
			content:  "my-secret-key\r\n",
			expected: "my-secret-key",
		},
		{
			name:     "trailing spaces and tabs",
			content:  "my-secret-key \t  ",
			expected: "my-secret-key",
		},
		{
			name:     "multiple trailing newlines and spaces",
			content:  "my-secret-key \r\n\n\t  \n",
			expected: "my-secret-key",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			filePath := filepath.Join(tmpDir, tc.name+".txt")
			if err := os.WriteFile(filePath, []byte(tc.content), 0600); err != nil {
				t.Fatalf("failed to write test key file: %v", err)
			}

			key, err := ResolveAPIKey("", filePath)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if key != tc.expected {
				t.Fatalf("expected '%s', got '%s'", tc.expected, key)
			}
		})
	}
}

func TestResolveAPIKey_Precedence(t *testing.T) {
	tmpDir := t.TempDir()
	keyPath := filepath.Join(tmpDir, "file_key.txt")

	if err := os.WriteFile(keyPath, []byte("file-key-value"), 0600); err != nil {
		t.Fatalf("failed to write temp key file: %v", err)
	}

	// When both inline key and key file are present, inline key must take precedence
	key, err := ResolveAPIKey("inline-key-value", keyPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if key != "inline-key-value" {
		t.Fatalf("expected 'inline-key-value', got '%s'", key)
	}

	// Even if keyFile points to a non-existent file, inline key must still succeed without error
	nonExistentPath := filepath.Join(tmpDir, "non_existent.txt")
	key, err = ResolveAPIKey("inline-key-value", nonExistentPath)
	if err != nil {
		t.Fatalf("unexpected error when inline key takes precedence over missing file: %v", err)
	}
	if key != "inline-key-value" {
		t.Fatalf("expected 'inline-key-value', got '%s'", key)
	}
}

func TestResolveAPIKey_MissingFile(t *testing.T) {
	nonExistentPath := filepath.Join(t.TempDir(), "does_not_exist.txt")

	key, err := ResolveAPIKey("", nonExistentPath)
	if err == nil {
		t.Fatalf("expected error for missing key file, got nil with key '%s'", key)
	}
	if key != "" {
		t.Fatalf("expected empty key on error, got '%s'", key)
	}
}

func TestResolveAPIKey_BothEmpty(t *testing.T) {
	key, err := ResolveAPIKey("", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if key != "" {
		t.Fatalf("expected empty key, got '%s'", key)
	}
}

func TestBackendFlagsRegistered(t *testing.T) {
	// Call Setup on all backend configurations to ensure flags are registered
	owm := &openWeatherConfig{}
	owm.Setup()

	caiyun := &CaiyunConfig{}
	caiyun.Setup()

	openmeteo := &openmeteoConfig{}
	openmeteo.Setup()

	pirate := &pirateweatherConfig{}
	pirate.Setup()

	weather := &weatherApiConfig{}
	weather.Setup()

	wwo := &wwoConfig{}
	wwo.Setup()

	expectedFlags := []string{
		"owm-api-key",
		"owm-api-key-file",
		"caiyun-api-key",
		"caiyun-api-key-file",
		"openmeteo-api-key",
		"openmeteo-api-key-file",
		"pirateweather-api-key",
		"pirateweather-api-key-file",
		"weather-api-key",
		"weather-api-key-file",
		"wwo-api-key",
		"wwo-api-key-file",
	}

	for _, flagName := range expectedFlags {
		f := flag.Lookup(flagName)
		if f == nil {
			t.Errorf("expected flag %s to be registered, but was nil", flagName)
		}
	}
}
