package rating

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"libriary_system/shared/domain"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (fn roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return fn(request)
}

func TestClient(t *testing.T) {
	for _, check := range []struct {
		method       string
		responseCode int
		responseBody string
		wantError    error
	}{
		{http.MethodGet, http.StatusOK, `{"stars":50}`, nil},
		{http.MethodPost, http.StatusCreated, "", nil},
		{http.MethodPut, http.StatusOK, `{"stars":60}`, nil},
		{http.MethodPatch, http.StatusOK, `{"stars":65}`, nil},
		{http.MethodGet, http.StatusNotFound, `{"message":"rating not found"}`, domain.ErrRatingNotFound},
		{http.MethodPost, http.StatusConflict, `{"message":"rating already exists"}`, domain.ErrRatingAlreadyExists},
	} {
		t.Run(check.method+"/"+http.StatusText(check.responseCode), func(t *testing.T) {
			transport := roundTripFunc(func(request *http.Request) (*http.Response, error) {
				if request.Method != check.method || request.URL.Path != "/api/v1/rating" || request.Header.Get("X-User-Name") != "alice" {
					t.Errorf("unexpected request: %s %s %s", request.Method, request.URL.Path, request.Header.Get("X-User-Name"))
				}
				return &http.Response{StatusCode: check.responseCode, Body: io.NopCloser(strings.NewReader(check.responseBody)), Header: make(http.Header)}, nil
			})
			client, err := NewClient("http://rating-service:8050", &http.Client{Transport: transport})
			if err != nil {
				t.Fatal(err)
			}
			ctx := context.Background()
			var actualErr error
			switch check.method {
			case http.MethodGet:
				_, actualErr = client.GetByUsername(ctx, "alice")
			case http.MethodPost:
				actualErr = client.Create(ctx, domain.Rating{Username: "alice", StarsCount: 50})
			case http.MethodPut:
				_, actualErr = client.UpdateRating(ctx, domain.Rating{Username: "alice", StarsCount: 60})
			case http.MethodPatch:
				_, actualErr = client.AddStars(ctx, "alice", 5)
			}
			if !errors.Is(actualErr, check.wantError) {
				t.Fatalf("error = %v, want %v", actualErr, check.wantError)
			}
		})
	}
}

func TestNewClientRejectsInvalidURL(t *testing.T) {
	if _, err := NewClient("rating-service:8050", nil); err == nil {
		t.Fatal("expected invalid URL error")
	}
}
