package supabase

import (
	"os"

	"github.com/nedpals/supabase-go"
)

func NewClient() *supabase.Client {

	supabaseURL := os.Getenv("SUPABASE_URL")
	supabaseKey := os.Getenv("SUPABASE_SERVICE_KEY")
	return supabase.CreateClient(supabaseURL, supabaseKey)
}
