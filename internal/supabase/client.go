package supabase

import (
	"os"

	"github.com/nedpals/supabase-go"
)

func NewClient() *supabase.Client {
	// Pro-tip: Use os.Getenv once we set up your .env file
	supabaseURL := os.Getenv("SUPABASE_URL")
	supabaseKey := os.Getenv("SUPABASE_ANON_KEY")
	return supabase.CreateClient(supabaseURL, supabaseKey)
}
