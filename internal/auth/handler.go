package auth

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/nedpals/supabase-go"
)

type Handler struct {
	Supabase        *supabase.Client
	PlacesAPIKey    string
	cacheMu         sync.Mutex
	restaurantCache map[string]restaurantCacheEntry
}

type restaurantCacheEntry struct {
	expiresAt time.Time
	results   []restaurantPlace
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
	googlePlacesEndpoint       = "https://maps.googleapis.com/maps/api/place/textsearch/json"
	googlePlacesNearbyEndpoint = "https://maps.googleapis.com/maps/api/place/nearbysearch/json"
	googlePlacesPhotoEndpoint  = "https://maps.googleapis.com/maps/api/place/photo"
	defaultPlacesQuery         = "restaurants in Plano Piloto, Brasilia"
	placesRequestTimeout       = 8 * time.Second
	nextPageTokenDelay         = 2 * time.Second
	maxPlacesPages             = 3
)

type nearbySearchSeed struct {
	Name   string
	Lat    float64
	Lng    float64
	Radius int
}

var planoPilotoNearbySeeds = []nearbySearchSeed{
	{Name: "Lago Norte", Lat: -15.7075, Lng: -47.8676, Radius: 3500},
	{Name: "Lago Sul", Lat: -15.8423, Lng: -47.8784, Radius: 3500},
	{Name: "Asa Norte", Lat: -15.7585, Lng: -47.8837, Radius: 3500},
	{Name: "Asa Sul", Lat: -15.7992, Lng: -47.9013, Radius: 3500},
	{Name: "Noroeste", Lat: -15.7428, Lng: -47.8922, Radius: 3000},
	{Name: "Sudoeste", Lat: -15.8046, Lng: -47.9218, Radius: 3000},
}

var planoPilotoNearbyKeywords = []string{
	"restaurant",
	"restaurante",
	"pizza",
	"sushi",
	"hamburgueria",
	"churrascaria",
	"cafe",
	"brunch",
}

var planoPilotoSupplementalQueries = []string{
	"restaurants in Plano Piloto, Brasilia",
	"restaurantes em Plano Piloto, Brasilia",
	"restaurants in Brasilia",
	"restaurantes em Brasilia",
}

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
	Results       []placeResult `json:"results"`
	Status        string        `json:"status"`
	ErrorMessage  string        `json:"error_message"`
	NextPageToken string        `json:"next_page_token"`
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
	explicitPageToken := strings.TrimSpace(r.URL.Query().Get("pageToken"))
	forceRefreshCache := strings.EqualFold(strings.TrimSpace(r.URL.Query().Get("refresh")), "1") ||
		strings.EqualFold(strings.TrimSpace(r.URL.Query().Get("refreshCache")), "1")

	var (
		results []restaurantPlace
		err     error
	)

	if queryText == "" && explicitPageToken == "" && !forceRefreshCache {
		if cached, ok := h.getCachedRestaurants("plano-piloto-map"); ok {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(cached)
			return
		}
	}

	if queryText != "" || explicitPageToken != "" {
		if queryText == "" {
			queryText = defaultPlacesQuery
		}
		results, err = h.fetchRestaurantsForQuery(r.Context(), queryText, explicitPageToken)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}
	} else {
		results, err = h.fetchPlanoPilotoRestaurants(r.Context())
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}
		h.setCachedRestaurants("plano-piloto-map", results, 5*time.Minute)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

func (h *Handler) fetchPlanoPilotoRestaurants(ctx context.Context) ([]restaurantPlace, error) {
	type fetchResult struct {
		places []restaurantPlace
		err    error
	}

	type task struct {
		name string
		fn   func() ([]restaurantPlace, error)
	}

	tasks := make([]task, 0, len(planoPilotoNearbySeeds)*len(planoPilotoNearbyKeywords)+len(planoPilotoSupplementalQueries))
	for _, seed := range planoPilotoNearbySeeds {
		for _, keyword := range planoPilotoNearbyKeywords {
			seedCopy := seed
			keywordCopy := keyword
			tasks = append(tasks, task{
				name: seedCopy.Name + ":" + keywordCopy,
				fn: func() ([]restaurantPlace, error) {
					return h.fetchNearbyRestaurantsPage(ctx, seedCopy, keywordCopy)
				},
			})
		}
	}

	for _, areaQuery := range planoPilotoSupplementalQueries {
		queryCopy := areaQuery
		tasks = append(tasks, task{
			name: queryCopy,
			fn: func() ([]restaurantPlace, error) {
				return h.fetchRestaurantsForQuery(ctx, queryCopy, "")
			},
		})
	}

	resultsCh := make(chan fetchResult, len(tasks))
	var wg sync.WaitGroup
	sem := make(chan struct{}, 6)

	for _, currentTask := range tasks {
		wg.Add(1)
		go func(currentTask task) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			places, err := currentTask.fn()
			resultsCh <- fetchResult{places: places, err: err}
		}(currentTask)
	}

	go func() {
		wg.Wait()
		close(resultsCh)
	}()

	aggregated := make([]restaurantPlace, 0, 240)
	seenByID := make(map[string]struct{}, 240)
	successfulQueries := 0

	for result := range resultsCh {
		if result.err != nil {
			log.Printf("Google Places query failed: %v", result.err)
			continue
		}

		successfulQueries++
		appendUniquePlaces(&aggregated, seenByID, result.places)
	}

	if successfulQueries == 0 {
		return nil, &placesHTTPError{}
	}

	return aggregated, nil
}

func (h *Handler) fetchRestaurantsForQuery(ctx context.Context, queryText string, explicitPageToken string) ([]restaurantPlace, error) {
	if strings.TrimSpace(queryText) == "" {
		queryText = defaultPlacesQuery
	}

	nextToken := explicitPageToken
	results := make([]restaurantPlace, 0, 45)

	for page := 0; page < maxPlacesPages; page++ {
		apiResponse, err := h.fetchPlacesPage(ctx, queryText, nextToken)
		if err != nil {
			return nil, err
		}

		switch apiResponse.Status {
		case "OK":
			// proceed
		case "ZERO_RESULTS":
			return results, nil
		default:
			log.Printf("Google Places API error: status=%s message=%s", apiResponse.Status, apiResponse.ErrorMessage)
			return nil, &placesHTTPError{}
		}

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

		if explicitPageToken != "" {
			break
		}

		nextToken = strings.TrimSpace(apiResponse.NextPageToken)
		if nextToken == "" {
			break
		}

		// Google may take a moment to activate next_page_token.
		select {
		case <-time.After(nextPageTokenDelay):
		case <-ctx.Done():
			return nil, context.DeadlineExceeded
		}
	}

	return results, nil
}

func (h *Handler) fetchNearbyRestaurantsPage(ctx context.Context, seed nearbySearchSeed, keyword string) ([]restaurantPlace, error) {
	requestCtx, cancel := context.WithTimeout(ctx, placesRequestTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(requestCtx, http.MethodGet, googlePlacesNearbyEndpoint, nil)
	if err != nil {
		return nil, err
	}

	params := req.URL.Query()
	params.Set("key", h.PlacesAPIKey)
	params.Set("location", formatLocation(seed.Lat, seed.Lng))
	params.Set("radius", formatRadius(seed.Radius))
	params.Set("type", "restaurant")
	if strings.TrimSpace(keyword) != "" {
		params.Set("keyword", keyword)
	}
	req.URL.RawQuery = params.Encode()

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Google Places HTTP error: status=%d", resp.StatusCode)
		return nil, &placesHTTPError{}
	}

	var apiResponse placesResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
		return nil, err
	}

	switch apiResponse.Status {
	case "OK":
		// proceed
	case "ZERO_RESULTS":
		return []restaurantPlace{}, nil
	default:
		log.Printf("Google Places API error: status=%s message=%s", apiResponse.Status, apiResponse.ErrorMessage)
		return nil, &placesHTTPError{}
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

	return results, nil
}

func appendUniquePlaces(aggregated *[]restaurantPlace, seenByID map[string]struct{}, places []restaurantPlace) {
	for _, place := range places {
		if place.ID != "" {
			if _, exists := seenByID[place.ID]; exists {
				continue
			}
			seenByID[place.ID] = struct{}{}
		}
		*aggregated = append(*aggregated, place)
	}
}

func formatLocation(lat float64, lng float64) string {
	return strings.TrimSpace(strings.Join([]string{trimFloat(lat), trimFloat(lng)}, ","))
}

func formatRadius(radius int) string {
	return strings.TrimSpace(trimInt(radius))
}

func trimFloat(value float64) string {
	return strings.TrimRight(strings.TrimRight(strings.TrimSpace(strconv.FormatFloat(value, 'f', 6, 64)), "0"), ".")
}

func trimInt(value int) string {
	return strconv.Itoa(value)
}

func (h *Handler) getCachedRestaurants(key string) ([]restaurantPlace, bool) {
	h.cacheMu.Lock()
	defer h.cacheMu.Unlock()

	if h.restaurantCache == nil {
		return nil, false
	}

	entry, ok := h.restaurantCache[key]
	if !ok || time.Now().After(entry.expiresAt) {
		return nil, false
	}

	results := make([]restaurantPlace, len(entry.results))
	copy(results, entry.results)
	return results, true
}

func (h *Handler) setCachedRestaurants(key string, results []restaurantPlace, ttl time.Duration) {
	h.cacheMu.Lock()
	defer h.cacheMu.Unlock()

	if h.restaurantCache == nil {
		h.restaurantCache = make(map[string]restaurantCacheEntry)
	}

	cachedResults := make([]restaurantPlace, len(results))
	copy(cachedResults, results)
	h.restaurantCache[key] = restaurantCacheEntry{
		expiresAt: time.Now().Add(ttl),
		results:   cachedResults,
	}
}

func (h *Handler) fetchPlacesPage(ctx context.Context, queryText string, pageToken string) (placesResponse, error) {
	requestCtx, cancel := context.WithTimeout(ctx, placesRequestTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(requestCtx, http.MethodGet, googlePlacesEndpoint, nil)
	if err != nil {
		return placesResponse{}, err
	}

	params := req.URL.Query()
	params.Set("key", h.PlacesAPIKey)
	if strings.TrimSpace(pageToken) != "" {
		params.Set("pagetoken", pageToken)
	} else {
		params.Set("query", queryText)
	}
	req.URL.RawQuery = params.Encode()

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return placesResponse{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Google Places HTTP error: status=%d", resp.StatusCode)
		return placesResponse{}, &placesHTTPError{}
	}

	var apiResponse placesResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
		return placesResponse{}, err
	}

	return apiResponse, nil
}

type placesHTTPError struct{}

func (e *placesHTTPError) Error() string {
	return "Google Places API error"
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
