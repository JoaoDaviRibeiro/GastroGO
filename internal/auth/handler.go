package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/nedpals/supabase-go"
)

// Handler holds our Supabase client dependency
type Handler struct {
	Supabase *supabase.Client
}

// Restaurant represents the structure of our database table
type Restaurant struct {
	ID        int64   `json:"id"`
	CreatedAt string  `json:"created_at"`
	Name      string  `json:"name"`
	Cuisine   string  `json:"cuisine"`
	Address   string  `json:"address"`
	Rating    float64 `json:"rating"`
	Lat       float64 `json:"lat"`
	Lng       float64 `json:"lng"`
	UserID    string  `json:"user_id"`
}

// AuthRequest is the "shape" of the JSON we expect from the frontend
type AuthRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// contextKey is a custom type to prevent context collisions
type contextKey string

const userKey contextKey = "user"

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

// Login handles authenticating existing users
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req AuthRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// This is the call to Supabase Auth
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

// IsAuthenticated is our middleware that guards protected routes
func (h *Handler) IsAuthenticated(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 1. Get the Authorization header
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
