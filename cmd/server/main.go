package main

import (
	"log"
	"net/http"

	// FIXED: Updated both internal imports
	"github.com/JoaoDaviRibeiro/GastroGO/internal/auth"
	"github.com/JoaoDaviRibeiro/GastroGO/internal/supabase"
)

func main() {
	sbClient := supabase.NewClient()
	authHandler := &auth.Handler{Supabase: sbClient}

	http.Handle("/", http.FileServer(http.Dir("./public")))
	http.HandleFunc("/api/signup", authHandler.SignUp)
	http.HandleFunc("/api/login", authHandler.Login)

	log.Println("Server running on :8080")
	http.ListenAndServe(":8080", nil)
}
