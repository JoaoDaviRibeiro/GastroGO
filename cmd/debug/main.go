package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/JoaoDaviRibeiro/GastroGO/internal/auth"
	"github.com/JoaoDaviRibeiro/GastroGO/internal/supabase"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()
	placesKey := os.Getenv("GOOGLE_PLACES_API_KEY")
	handler := &auth.Handler{Supabase: supabase.NewClient(), PlacesAPIKey: placesKey}

	results, err := handler.FetchForDebug("restaurants in Brasilia")
	if err != nil {
		log.Printf("fetch failed: %v", err)
		// Try direct HTTP call for diagnostics
		endpoint := "https://maps.googleapis.com/maps/api/place/textsearch/json"
		params := url.Values{}
		params.Set("query", "restaurants in Brasilia")
		params.Set("key", placesKey)
		resp, rerr := http.Get(endpoint + "?" + params.Encode())
		if rerr != nil {
			log.Fatalf("direct request failed: %v", rerr)
		}
		defer resp.Body.Close()
		body, _ := ioutil.ReadAll(resp.Body)
		fmt.Println("--- Direct response status:", resp.Status)
		fmt.Println(string(body))
		os.Exit(1)
	}

	out, _ := json.MarshalIndent(results, "", "  ")
	fmt.Println(string(out))
}
