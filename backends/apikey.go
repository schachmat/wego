package backends

import (
	"fmt"
	"os"
	"strings"
)

// ResolveAPIKey determines the API key to use based on key and keyFile.
// If key is non-empty, it takes precedence and is returned directly.
// If key is empty and keyFile is non-empty, the contents of keyFile are read
// with any trailing whitespace and newlines trimmed.
// If both key and keyFile are empty, an empty string and nil error are returned.
func ResolveAPIKey(key, keyFile string) (string, error) {
	if key != "" {
		return key, nil
	}
	if keyFile == "" {
		return "", nil
	}
	data, err := os.ReadFile(keyFile)
	if err != nil {
		return "", fmt.Errorf("unable to read api key file %s: %w", keyFile, err)
	}
	return strings.TrimRight(string(data), "\r\n \t"), nil
}
