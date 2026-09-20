package library

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"

	"libriary_system/shared/domain"
	"libriary_system/shared/pagination"
)

func TestClientListLibrariesByCity(t *testing.T) {
	libraryUID := uuid.MustParse("83575e12-7ce0-48ee-9931-51919ff3c9ee")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		if request.URL.Path != "/api/v1/libraries" {
			t.Fatalf("path = %q", request.URL.Path)
		}
		query := request.URL.Query()
		if query.Get("city") != "Москва" || query.Get("limit") != "10" || query.Get("offset") != "20" {
			t.Fatalf("query = %v", query)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"items":[{
				"libraryUid":"83575e12-7ce0-48ee-9931-51919ff3c9ee",
				"name":"Библиотека",
				"city":"Москва",
				"address":"Адрес"
			}],
			"total":21
		}`))
	}))
	defer server.Close()
	client, err := NewClient(server.URL, server.Client())
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	page, err := client.ListLibrariesByCity(
		context.Background(),
		"Москва",
		pagination.PageToken{Limit: 10, Offset: 20},
	)
	if err != nil {
		t.Fatalf("ListLibrariesByCity() error = %v", err)
	}
	if page.Total != 21 || len(page.Items) != 1 || page.Items[0].Id != libraryUID {
		t.Fatalf("page = %+v", page)
	}
}

func TestClientReserveBookMapsUnavailableError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		_, _ = w.Write([]byte(`{"message":"book is not available"}`))
	}))
	defer server.Close()
	client, err := NewClient(server.URL, server.Client())
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	_, err = client.ReserveBook(context.Background(), uuid.New(), uuid.New())
	if !errors.Is(err, domain.ErrBookUnavailable) {
		t.Fatalf("ReserveBook() error = %v, want ErrBookUnavailable", err)
	}
}
