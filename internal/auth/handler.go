package auth

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/nedpals/supabase-go"
)

type Handler struct {
	Supabase *supabase.Client
}

// ScoreData matches the nested structure returned by our SQL View join
type ScoreData struct {
	AverageScore float64 `json:"average_score"`
	TotalReviews int     `json:"total_reviews"`
}

type Restaurant struct {
	ID               int64       `json:"id"`
	CreatedAt        string      `json:"created_at"`
	Name             string      `json:"name"`
	Cuisine          string      `json:"cuisine"`
	Address          string      `json:"address"`
	Lat              float64     `json:"lat"`
	Lng              float64     `json:"lng"`
	UserID           string      `json:"user_id"`
	RestaurantScores []ScoreData `json:"restaurant_scores"`
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
	var results []Restaurant

	// The supabase-go client uses the PostgREST syntax
	err := h.Supabase.DB.From("restaurants").Select("*").Execute(&results)
	if err != nil {
		http.Error(w, "Database error: "+err.Error(), http.StatusInternalServerError)
		return
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
