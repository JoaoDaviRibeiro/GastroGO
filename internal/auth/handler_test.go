package auth

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
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
	want := "https://maps.googleapis.com/maps/api/place/photo?key=api-key&maxwidth=800&photo_reference=photo-ref"
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
		if got := req.URL.Query().Get("query"); got != "custom query" {
			t.Fatalf("unexpected query: %s", got)
		}
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

	expectedPhoto := "https://maps.googleapis.com/maps/api/place/photo?key=test-key&maxwidth=800&photo_reference=ref123"
	if got.PhotoURL != expectedPhoto {
		t.Fatalf("unexpected photo url: %s", got.PhotoURL)
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
