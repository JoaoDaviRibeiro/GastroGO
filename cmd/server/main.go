package main

import (
	"log"
	"net/http"

	// FIXED: Updated both internal
	"github.com/JoaoDaviRibeiro/GastroGO/internal/auth"
	"github.com/JoaoDaviRibeiro/GastroGO/internal/supabase"
	"github.com/joho/godotenv"
)

func main() {
	// 1. Load Environment Variables
	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file, checking system env...")
	}

	// 2. Initialize Supabase Client and Auth Handler
	sbClient := supabase.NewClient()
	authHandler := &auth.Handler{Supabase: sbClient}

	// 3. Static File Server (Frontend)
	http.Handle("/", http.FileServer(http.Dir("./public")))

	// 4. Public API Routes
	http.HandleFunc("/api/signup", authHandler.SignUp)
	http.HandleFunc("/api/login", authHandler.Login)

	// 5. Protected API Routes
	// We wrap the Dashboard handler with IsAuthenticated middleware
	http.HandleFunc("/api/dashboard", authHandler.IsAuthenticated(authHandler.Dashboard))

	log.Println("GastroGO server running on :8080")
	// Using log.Fatal to ensure we catch any startup crashes
	log.Fatal(http.ListenAndServe(":8080", nil))
}
