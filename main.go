package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
)

const (
	PORT            = 8080
	MAPBOX_BASE_URL = "https://api.mapbox.com"
)

type LocationRequest struct {
	From string `json:"from"`
	To   string `json:"to"`
}

type DistanceResponse struct {
	Distance     float64   `json:"distance"`
	Duration     float64   `json:"duration"`
	FromCoords   []float64 `json:"from_coords"`
	ToCoords     []float64 `json:"to_coords"`
	GeometryLine string    `json:"geometry_line,omitempty"`
}

type MapboxGeocodingResponse struct {
	Features []struct {
		Center []float64 `json:"center"`
	} `json:"features"`
}

type MapboxDirectionsResponse struct {
	Routes []struct {
		Distance float64 `json:"distance"`
		Duration float64 `json:"duration"`
		Geometry string  `json:"geometry"`
	} `json:"routes"`
}

func enableCORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next(w, r)
	}
}

func main() {
	apiKey := os.Getenv("MAPBOX_API_KEY")
	if apiKey == "" {
		log.Fatal("MAPBOX_API_KEY environment variable is required")
	}

	http.HandleFunc("/distance", enableCORS(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req LocationRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		if req.From == "" || req.To == "" {
			http.Error(w, "Both 'from' and 'to' addresses are required", http.StatusBadRequest)
			return
		}

		fromCoords, err := geocodeAddress(req.From, apiKey)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to geocode 'from' address: %v", err), http.StatusInternalServerError)
			return
		}

		toCoords, err := geocodeAddress(req.To, apiKey)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to geocode 'to' address: %v", err), http.StatusInternalServerError)
			return
		}

		distance, duration, geometry, err := getDirections(fromCoords, toCoords, apiKey)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to get directions: %v", err), http.StatusInternalServerError)
			return
		}

		resp := DistanceResponse{
			Distance:     distance / 1000,
			Duration:     duration / 60,
			FromCoords:   fromCoords,
			ToCoords:     toCoords,
			GeometryLine: geometry,
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
			return
		}
	}))

	http.HandleFunc("/health", enableCORS(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))

	log.Printf("Server starting on port %d...\n", PORT)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", PORT), nil); err != nil {
		log.Fatal(err)
	}
}

func geocodeAddress(address string, apiKey string) ([]float64, error) {
	encodedAddress := url.QueryEscape(address)
	geocodingURL := fmt.Sprintf("%s/geocoding/v5/mapbox.places/%s.json?access_token=%s",
		MAPBOX_BASE_URL, encodedAddress, apiKey)

	resp, err := http.Get(geocodingURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("geocoding API returned status %d: %s", resp.StatusCode, string(body))
	}

	var geocodingResp MapboxGeocodingResponse
	if err := json.Unmarshal(body, &geocodingResp); err != nil {
		return nil, err
	}

	if len(geocodingResp.Features) == 0 {
		return nil, fmt.Errorf("no results found for address: %s", address)
	}

	return geocodingResp.Features[0].Center, nil
}

func getDirections(fromCoords, toCoords []float64, apiKey string) (float64, float64, string, error) {
	directionsURL := fmt.Sprintf("%s/directions/v5/mapbox/driving/%f,%f;%f,%f?access_token=%s&geometries=polyline",
		MAPBOX_BASE_URL,
		fromCoords[0], fromCoords[1],
		toCoords[0], toCoords[1],
		apiKey)

	resp, err := http.Get(directionsURL)
	if err != nil {
		return 0, 0, "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, 0, "", err
	}

	if resp.StatusCode != http.StatusOK {
		return 0, 0, "", fmt.Errorf("directions API returned status %d: %s", resp.StatusCode, string(body))
	}

	var directionsResp MapboxDirectionsResponse
	if err := json.Unmarshal(body, &directionsResp); err != nil {
		return 0, 0, "", err
	}

	if len(directionsResp.Routes) == 0 {
		return 0, 0, "", fmt.Errorf("no route found between the specified locations")
	}

	route := directionsResp.Routes[0]
	return route.Distance, route.Duration, route.Geometry, nil
}
