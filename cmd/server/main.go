package main

import (
	"log"
	"net/http"

	"github.com/JoaoDaviRibeiro/GastroGO/internal/auth"
	"github.com/JoaoDaviRibeiro/GastroGO/internal/supabase"
	"github.com/joho/godotenv"
)

func main() {
	// 1. Load Environment Variables
	// This pulls your SUPABASE_URL and SUPABASE_KEY
	err := godotenv.Load()
	if err != nil {
		log.Println("Note: .env file not found, using system environment variables")
	}

	// 2. Initialize Supabase Client and Auth Handler
	sbClient := supabase.NewClient()
	authHandler := &auth.Handler{Supabase: sbClient}

	// 3. Static File Server (Frontend)
	// Serves index.html, dashboard.html, and script.js from the public folder
	http.Handle("/", http.FileServer(http.Dir("./public")))

	// 4. Public API Routes
	// No token needed for these
	http.HandleFunc("/api/signup", authHandler.SignUp)
	http.HandleFunc("/api/login", authHandler.Login)

	// 5. Protected API Routes
	// These all require a valid Bearer Token in the Authorization header

	// Get user dashboard info
	http.HandleFunc("/api/dashboard", authHandler.IsAuthenticated(authHandler.Dashboard))

	// Get the list of all restaurants
	http.HandleFunc("/api/restaurants", authHandler.IsAuthenticated(authHandler.GetRestaurants))

	// Submit a Letterboxd-style rating
	http.HandleFunc("/api/rate", authHandler.IsAuthenticated(authHandler.RateRestaurant))

	// 6. Start the Server
	log.Println("🚀 GastroGO server is live on http://localhost:8080")

	// log.Fatal will log the error and exit if the port is already in use
	log.Fatal(http.ListenAndServe(":8080", nil))
}
