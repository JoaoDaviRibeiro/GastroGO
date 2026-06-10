package auth

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func stubHTTPClient(t *testing.T, rt http.RoundTripper) {
	t.Helper()
	original := http.DefaultClient
	http.DefaultClient = &http.Client{Transport: rt}
	t.Cleanup(func() {
		http.DefaultClient = original
	})
}

func TestFirstNonEmpty(t *testing.T) {
	cases := []struct {
		name     string
		input    []string
		expected string
	}{
		{"all empty", []string{"", "   "}, ""},
		{"trims spaces", []string{"  foo  ", "bar"}, "foo"},
		{"first match later", []string{"", "bar"}, "bar"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := firstNonEmpty(tc.input...)
			if got != tc.expected {
				t.Fatalf("expected %q got %q", tc.expected, got)
			}
		})
	}
}

func TestBuildPhotoURL(t *testing.T) {
	photos := []placePhoto{{PhotoReference: "photo-ref", Width: 400, Height: 300}}
	url := buildPhotoURL(photos, "api-key")
	want := "/api/restaurants/photo?maxwidth=800&ref=photo-ref"
	if url != want {
		t.Fatalf("expected %q got %q", want, url)
	}

	if got := buildPhotoURL(nil, "api-key"); got != "" {
		t.Fatalf("expected empty url got %q", got)
	}

	if got := buildPhotoURL(photos, ""); got != "" {
		t.Fatalf("expected empty url when api key missing got %q", got)
	}
}

func TestGetRestaurantsSuccess(t *testing.T) {
	stubHTTPClient(t, roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if got := req.URL.Query().Get("key"); got != "test-key" {
			t.Fatalf("missing api key, got %s", got)
		}
		if got := req.URL.Query().Get("pagetoken"); got != "token123" {
			t.Fatalf("missing page token, got %s", got)
		}

		body := `{"results":[{"place_id":"abc","name":"Resto","formatted_address":"Av. Paulista","vicinity":"","rating":4.7,"user_ratings_total":87,"geometry":{"location":{"lat":-23.5,"lng":-46.6}},"photos":[{"photo_reference":"ref123","width":800,"height":600}],"types":["restaurant","food"]}],"status":"OK"}`
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(body)),
			Header:     make(http.Header),
		}, nil
	}))

	handler := &Handler{PlacesAPIKey: "test-key"}
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/restaurants?query=custom+query&pageToken=token123", nil)

	handler.GetRestaurants(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200 got %d", rr.Code)
	}

	var payload []restaurantPlace
	if err := json.Unmarshal(rr.Body.Bytes(), &payload); err != nil {
		t.Fatalf("failed to decode payload: %v", err)
	}

	if len(payload) != 1 {
		t.Fatalf("expected 1 result got %d", len(payload))
	}

	got := payload[0]
	if got.ID != "abc" || got.Name != "Resto" {
		t.Fatalf("unexpected place data: %+v", got)
	}

	expectedPhoto := "/api/restaurants/photo?maxwidth=800&ref=ref123"
	if got.PhotoURL != expectedPhoto {
		t.Fatalf("unexpected photo url: %s", got.PhotoURL)
	}
}

func TestGetRestaurantsAggregatesMultiplePages(t *testing.T) {
	callCount := 0
	stubHTTPClient(t, roundTripFunc(func(req *http.Request) (*http.Response, error) {
		callCount++
		token := req.URL.Query().Get("pagetoken")
		switch callCount {
		case 1:
			if token != "" {
				t.Fatalf("expected no page token on first call, got %q", token)
			}
			body := `{"results":[`
			for i := 1; i <= 15; i++ {
				if i > 1 {
					body += `,`
				}
				body += `{"place_id":"id` + string(rune('A'+i-1)) + `","name":"R` + string(rune('A'+i-1)) + `","formatted_address":"Addr","rating":4.1,"user_ratings_total":10,"geometry":{"location":{"lat":-15.0,"lng":-47.0}},"types":["restaurant"]}`
			}
			body += `],"status":"OK","next_page_token":"next-token"}`
			return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(body)), Header: make(http.Header)}, nil
		case 2:
			if token != "next-token" {
				t.Fatalf("expected next page token on second call, got %q", token)
			}
			body := `{"results":[{"place_id":"idZ","name":"RZ","formatted_address":"Addr","rating":4.3,"user_ratings_total":22,"geometry":{"location":{"lat":-15.1,"lng":-47.1}},"types":["restaurant"]}],"status":"OK"}`
			return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(body)), Header: make(http.Header)}, nil
		default:
			t.Fatalf("unexpected extra call: %d", callCount)
			return nil, nil
		}
	}))

	handler := &Handler{PlacesAPIKey: "test-key"}
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/restaurants?query=custom+query", nil)

	handler.GetRestaurants(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200 got %d", rr.Code)
	}

	var payload []restaurantPlace
	if err := json.Unmarshal(rr.Body.Bytes(), &payload); err != nil {
		t.Fatalf("failed to decode payload: %v", err)
	}

	if len(payload) != 16 {
		t.Fatalf("expected 16 results after aggregation got %d", len(payload))
	}
}

func TestGetRestaurantsZeroResults(t *testing.T) {
	stubHTTPClient(t, roundTripFunc(func(req *http.Request) (*http.Response, error) {
		body := `{"results":[],"status":"ZERO_RESULTS"}`
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(body)),
			Header:     make(http.Header),
		}, nil
	}))

	handler := &Handler{PlacesAPIKey: "key"}
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/restaurants", nil)

	handler.GetRestaurants(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200 got %d", rr.Code)
	}

	if strings.TrimSpace(rr.Body.String()) != "[]" {
		t.Fatalf("expected empty json array got %s", rr.Body.String())
	}
}

func TestGetRestaurantsMissingKey(t *testing.T) {
	handler := &Handler{}
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/restaurants", nil)

	handler.GetRestaurants(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500 got %d", rr.Code)
	}
}

func TestGetRestaurantsPlanoPilotoAreasAggregatedAndDeduped(t *testing.T) {
	seenQueries := map[string]int{}
	var seenMu sync.Mutex
	stubHTTPClient(t, roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if req.URL.Query().Get("pagetoken") != "" {
			t.Fatalf("unexpected page token for default mapping fetch")
		}

		if keyword := req.URL.Query().Get("keyword"); keyword != "" {
			query := keyword + "|" + req.URL.Query().Get("location")
			seenMu.Lock()
			seenQueries[query]++
			seenMu.Unlock()

			switch query {
			case "restaurant|-15.7075,-47.8676":
				body := `{"results":[{"place_id":"dup-1","name":"A","formatted_address":"Addr A","rating":4.0,"user_ratings_total":10,"geometry":{"location":{"lat":-15.7,"lng":-47.8}},"types":["restaurant"]}],"status":"OK"}`
				return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(body)), Header: make(http.Header)}, nil
			case "pizza|-15.8423,-47.8784":
				body := `{"results":[{"place_id":"dup-1","name":"A","formatted_address":"Addr A","rating":4.0,"user_ratings_total":10,"geometry":{"location":{"lat":-15.7,"lng":-47.8}},"types":["restaurant"]},{"place_id":"unique-2","name":"B","formatted_address":"Addr B","rating":4.2,"user_ratings_total":20,"geometry":{"location":{"lat":-15.8,"lng":-47.9}},"types":["restaurant"]}],"status":"OK"}`
				return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(body)), Header: make(http.Header)}, nil
			default:
				body := `{"results":[],"status":"ZERO_RESULTS"}`
				return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(body)), Header: make(http.Header)}, nil
			}
		}

		body := `{"results":[],"status":"ZERO_RESULTS"}`
		return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(body)), Header: make(http.Header)}, nil
	}))

	handler := &Handler{PlacesAPIKey: "test-key"}
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/restaurants", nil)

	handler.GetRestaurants(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200 got %d", rr.Code)
	}

	if seenQueries["restaurant|-15.7075,-47.8676"] == 0 {
		t.Fatalf("expected nearby query for Lago Norte restaurant seed")
	}
	if seenQueries["pizza|-15.8423,-47.8784"] == 0 {
		t.Fatalf("expected nearby query for Lago Sul pizza seed")
	}

	var payload []restaurantPlace
	if err := json.Unmarshal(rr.Body.Bytes(), &payload); err != nil {
		t.Fatalf("failed to decode payload: %v", err)
	}

	if len(payload) != 2 {
		t.Fatalf("expected deduped 2 results got %d", len(payload))
	}
}

func TestGetPopularTimesBatch(t *testing.T) {
	handler := &Handler{
		popularTimesFetcher: func(ctx context.Context, placeID string) (popularTimesPayload, error) {
			current := 42
			return popularTimesPayload{
				ID:                placeID,
				CurrentPopularity: &current,
				PopularTimes: []popularTimesDay{{
					Name: "Monday",
					Data: []int{0, 10, 20, 42},
				}},
			}, nil
		},
	}

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/restaurants/popular-times?placeIds=abc,def", nil)

	handler.GetPopularTimes(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200 got %d", rr.Code)
	}

	var payload popularTimesBatchResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &payload); err != nil {
		t.Fatalf("failed to decode payload: %v", err)
	}

	if len(payload.Results) != 2 {
		t.Fatalf("expected 2 popular times entries got %d", len(payload.Results))
	}

	if payload.Results["abc"].ID != "abc" || payload.Results["def"].ID != "def" {
		t.Fatalf("unexpected popular times payload: %+v", payload.Results)
	}
}
