package rating

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"libriary_system/gateway_service/internal/domain"
	"libriary_system/gateway_service/internal/usecase"
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
		dependency   bool
	}{
		{method: http.MethodGet, responseCode: http.StatusOK, responseBody: `{"stars":50}`},
		{method: http.MethodPost, responseCode: http.StatusCreated},
		{method: http.MethodPatch, responseCode: http.StatusOK, responseBody: `{"stars":65}`},
		{method: http.MethodGet, responseCode: http.StatusNotFound, responseBody: `{"message":"rating not found"}`, wantError: domain.ErrRatingNotFound},
		{method: http.MethodPost, responseCode: http.StatusConflict, responseBody: `{"message":"rating already exists"}`, wantError: domain.ErrRatingAlreadyExists},
		{method: http.MethodGet, responseCode: http.StatusInternalServerError, responseBody: `{"message":"failed"}`, dependency: true},
	} {
		t.Run(check.method+"/"+http.StatusText(check.responseCode), func(t *testing.T) {
			transport := roundTripFunc(func(request *http.Request) (*http.Response, error) {
				if request.Method != check.method || request.URL.Path != "/api/v1/rating" || request.Header.Get("X-User-Name") != "alice" {
					t.Errorf("unexpected request: %s %s %s", request.Method, request.URL.Path, request.Header.Get("X-User-Name"))
				}
				return &http.Response{
					StatusCode: check.responseCode,
					Body:       io.NopCloser(strings.NewReader(check.responseBody)),
					Header:     make(http.Header),
				}, nil
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
			case http.MethodPatch:
				_, actualErr = client.AddStars(ctx, "alice", 5)
			}
			if check.wantError != nil && !errors.Is(actualErr, check.wantError) {
				t.Fatalf("error = %v, want %v", actualErr, check.wantError)
			}
			if check.wantError == nil && !check.dependency && actualErr != nil {
				t.Fatalf("unexpected error = %v", actualErr)
			}
			if check.dependency {
				var dependencyErr *usecase.DependencyError
				if !errors.As(actualErr, &dependencyErr) {
					t.Fatalf("error = %v, want DependencyError", actualErr)
				}
			}
		})
	}
}

func TestNewClientRejectsInvalidURL(t *testing.T) {
	if _, err := NewClient("rating-service:8050", nil); err == nil {
		t.Fatal("expected invalid URL error")
	}
}
