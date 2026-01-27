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

	// Load .env file
	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file, checking system env...")
	}

	sbClient := supabase.NewClient()
	authHandler := &auth.Handler{Supabase: sbClient}

	http.Handle("/", http.FileServer(http.Dir("./public")))
	http.HandleFunc("/api/signup", authHandler.SignUp)
	http.HandleFunc("/api/login", authHandler.Login)

	log.Println("Server running on :8080")
	http.ListenAndServe(":8080", nil)
}
