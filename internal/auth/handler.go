package auth

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/nedpals/supabase-go"
)

type Handler struct {
	Supabase     *supabase.Client
	PlacesAPIKey string
}

type ReviewRequest struct {
	RestaurantID int64   `json:"restaurant_id"`
	Rating       float64 `json:"rating"`
}

type AuthRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type contextKey string

const userKey contextKey = "user"

const (
	googlePlacesEndpoint      = "https://maps.googleapis.com/maps/api/place/textsearch/json"
	googlePlacesPhotoEndpoint = "https://maps.googleapis.com/maps/api/place/photo"
	defaultPlacesQuery        = "restaurants in Sao Paulo"
	placesRequestTimeout      = 8 * time.Second
)

type placePhoto struct {
	PhotoReference string `json:"photo_reference"`
	Width          int    `json:"width"`
	Height         int    `json:"height"`
}

type placeLocation struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

type placeGeometry struct {
	Location placeLocation `json:"location"`
}

type placeResult struct {
	PlaceID          string        `json:"place_id"`
	Name             string        `json:"name"`
	FormattedAddress string        `json:"formatted_address"`
	Vicinity         string        `json:"vicinity"`
	Rating           float64       `json:"rating"`
	UserRatingsTotal int           `json:"user_ratings_total"`
	Geometry         placeGeometry `json:"geometry"`
	Photos           []placePhoto  `json:"photos"`
	Types            []string      `json:"types"`
}

type placesResponse struct {
	Results      []placeResult `json:"results"`
	Status       string        `json:"status"`
	ErrorMessage string        `json:"error_message"`
}

type restaurantPlace struct {
	ID               string   `json:"id"`
	Name             string   `json:"name"`
	Address          string   `json:"address"`
	Rating           float64  `json:"rating"`
	UserRatingsTotal int      `json:"user_ratings_total"`
	PhotoURL         string   `json:"photo_url,omitempty"`
	Lat              float64  `json:"lat"`
	Lng              float64  `json:"lng"`
	Types            []string `json:"types"`
}

// RateRestaurant - Ensure this matches the name in main.go
func (h *Handler) RateRestaurant(w http.ResponseWriter, r *http.Request) {
	var req ReviewRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}

	val := r.Context().Value(userKey)
	user := val.(*supabase.User)

	err := h.Supabase.DB.From("reviews").Insert(map[string]interface{}{
		"restaurant_id": req.RestaurantID,
		"user_id":       user.ID,
		"rating":        req.Rating,
	}).Execute(nil)

	if err != nil {
		log.Println("DATABASE ERROR:", err.Error())
		http.Error(w, "Failed to save review", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

// --- Rest of your handlers (Login, SignUp, GetRestaurants, etc.) ---
// Make sure to include IsAuthenticated, Login, SignUp, Dashboard, and GetRestaurants below!
// --- PUBLIC HANDLERS ---

func (h *Handler) SignUp(w http.ResponseWriter, r *http.Request) {
	var req AuthRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	user, err := h.Supabase.Auth.SignUp(r.Context(), supabase.UserCredentials{
		Email:    req.Email,
		Password: req.Password,
	})

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req AuthRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	details, err := h.Supabase.Auth.SignIn(r.Context(), supabase.UserCredentials{
		Email:    req.Email,
		Password: req.Password,
	})

	if err != nil {
		http.Error(w, "Unauthorized: "+err.Error(), http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(details)
}

// --- MIDDLEWARE & PROTECTED HANDLERS ---

func (h *Handler) IsAuthenticated(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Authorization header required", http.StatusUnauthorized)
			return
		}

		// 2. Remove "Bearer " prefix to get just the JWT
		token := strings.TrimPrefix(authHeader, "Bearer ")

		// 3. Validate the token with Supabase
		user, err := h.Supabase.Auth.User(r.Context(), token)
		if err != nil {
			http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), userKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

// GetRestaurants fetches all data from the restaurants table
func (h *Handler) GetRestaurants(w http.ResponseWriter, r *http.Request) {
	if strings.TrimSpace(h.PlacesAPIKey) == "" {
		http.Error(w, "Google Places API key is not configured", http.StatusInternalServerError)
		return
	}

	queryText := strings.TrimSpace(r.URL.Query().Get("query"))
	if queryText == "" {
		queryText = defaultPlacesQuery
	}

	ctx, cancel := context.WithTimeout(r.Context(), placesRequestTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, googlePlacesEndpoint, nil)
	if err != nil {
		http.Error(w, "Failed to build Google Places request", http.StatusInternalServerError)
		return
	}

	params := req.URL.Query()
	params.Set("query", queryText)
	params.Set("key", h.PlacesAPIKey)
	if token := strings.TrimSpace(r.URL.Query().Get("pageToken")); token != "" {
		params.Set("pagetoken", token)
	}
	req.URL.RawQuery = params.Encode()

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		http.Error(w, "Google Places API request failed", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Google Places HTTP error: status=%d", resp.StatusCode)
		http.Error(w, "Google Places API error", http.StatusBadGateway)
		return
	}

	var apiResponse placesResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
		http.Error(w, "Failed to decode Google Places response", http.StatusBadGateway)
		return
	}

	switch apiResponse.Status {
	case "OK":
		// proceed
	case "ZERO_RESULTS":
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]restaurantPlace{})
		return
	default:
		log.Printf("Google Places API error: status=%s message=%s", apiResponse.Status, apiResponse.ErrorMessage)
		http.Error(w, "Google Places API error", http.StatusBadGateway)
		return
	}

	results := make([]restaurantPlace, 0, len(apiResponse.Results))
	for _, place := range apiResponse.Results {
		results = append(results, restaurantPlace{
			ID:               place.PlaceID,
			Name:             place.Name,
			Address:          firstNonEmpty(place.FormattedAddress, place.Vicinity),
			Rating:           place.Rating,
			UserRatingsTotal: place.UserRatingsTotal,
			PhotoURL:         buildPhotoURL(place.Photos, h.PlacesAPIKey),
			Lat:              place.Geometry.Location.Lat,
			Lng:              place.Geometry.Location.Lng,
			Types:            place.Types,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

func (h *Handler) Dashboard(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(userKey).(*supabase.User)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Welcome to the secret GastroGO Dashboard!",
		"email":   user.Email,
	})
}

func buildPhotoURL(photos []placePhoto, apiKey string) string {
	if len(photos) == 0 || strings.TrimSpace(apiKey) == "" {
		return ""
	}

	ref := strings.TrimSpace(photos[0].PhotoReference)
	if ref == "" {
		return ""
	}

	values := url.Values{}
	values.Set("maxwidth", "800")
	values.Set("photo_reference", ref)
	values.Set("key", apiKey)
	return googlePlacesPhotoEndpoint + "?" + values.Encode()
}

func firstNonEmpty(values ...string) string {
	for _, val := range values {
		if trimmed := strings.TrimSpace(val); trimmed != "" {
			return trimmed
		}
	}
	return ""
}
