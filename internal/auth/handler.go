package auth

import (
	"context" // Required for passing user data through middleware
	"encoding/json"
	"net/http"
	"strings" // Required for parsing the "Bearer" string

	"github.com/nedpals/supabase-go"
)

// Handler holds our Supabase client dependency
type Handler struct {
	Supabase *supabase.Client
}

// AuthRequest is the "shape" of the JSON we expect from the frontend
type AuthRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

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

		// 4. Store user in context (useful for personalized data later)
		ctx := context.WithValue(r.Context(), "user", user)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

// Dashboard is an example of a protected route
func (h *Handler) Dashboard(w http.ResponseWriter, r *http.Request) {
	// Retrieve the user from context (injected by the middleware)
	user := r.Context().Value("user").(*supabase.User)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Welcome to the secret GastroGO Dashboard!",
		"email":   user.Email,
	})
}
