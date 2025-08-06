package location

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// LocationProvider defines the interface for location services
type LocationProvider interface {
	GetLocation() (*GeoLocation, error)
}

// GeoLocation represents a geographical location
type GeoLocation struct {
	Latitude  float64 `json:"lat"`
	Longitude float64 `json:"lon"`
	City      string  `json:"city"`
	Country   string  `json:"country"`
	Provider  string  `json:"provider"`
}

// String returns the location as a coordinate string for wego
func (g *GeoLocation) String() string {
	return fmt.Sprintf("%.3f,%.3f", g.Latitude, g.Longitude)
}

// IPLocationProvider uses IP-based geolocation services
type IPLocationProvider struct {
	Timeout time.Duration
	APIKey  string // Optional API key for premium services
}

// NewIPLocationProvider creates a new IP location provider
func NewIPLocationProvider(timeout time.Duration) *IPLocationProvider {
	return &IPLocationProvider{
		Timeout: timeout,
	}
}

// ipApiResponse represents the response from ip-api.com
type ipApiResponse struct {
	Status  string  `json:"status"`
	City    string  `json:"city"`
	Country string  `json:"country"`
	Lat     float64 `json:"lat"`
	Lon     float64 `json:"lon"`
	Message string  `json:"message,omitempty"`
}

// GetLocation implements LocationProvider interface using ip-api.com (free service)
func (p *IPLocationProvider) GetLocation() (*GeoLocation, error) {
	client := &http.Client{Timeout: p.Timeout}
	
	// Use ip-api.com free service (15 requests per minute limit)
	resp, err := client.Get("http://ip-api.com/json/")
	if err != nil {
		return nil, fmt.Errorf("failed to fetch location: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var apiResp ipApiResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if apiResp.Status != "success" {
		return nil, fmt.Errorf("geolocation API error: %s", apiResp.Message)
	}

	return &GeoLocation{
		Latitude:  apiResp.Lat,
		Longitude: apiResp.Lon,
		City:      apiResp.City,
		Country:   apiResp.Country,
		Provider:  "ip-api.com",
	}, nil
}

// GetCurrentLocation is a convenience function to get current location using IP
func GetCurrentLocation() (*GeoLocation, error) {
	provider := NewIPLocationProvider(10 * time.Second)
	return provider.GetLocation()
}