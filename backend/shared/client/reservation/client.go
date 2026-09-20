package reservation

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"

	"libriary_system/shared/domain"
)

const maxResponseBodySize = 1 << 20

type Client struct {
	baseURL    *url.URL
	httpClient *http.Client
}

type ServiceError struct {
	StatusCode int
	Message    string
}

func (err *ServiceError) Error() string {
	return fmt.Sprintf("reservation service returned status %d: %s", err.StatusCode, err.Message)
}

func NewClient(rawBaseURL string, httpClient *http.Client) (*Client, error) {
	baseURL, err := url.Parse(strings.TrimSpace(rawBaseURL))
	if err != nil {
		return nil, fmt.Errorf("parse reservation service URL: %w", err)
	}
	if baseURL.Scheme != "http" && baseURL.Scheme != "https" {
		return nil, fmt.Errorf("reservation service URL must use http or https scheme")
	}
	if baseURL.Host == "" {
		return nil, fmt.Errorf("reservation service URL must contain a host")
	}
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 5 * time.Second}
	}
	return &Client{baseURL: baseURL, httpClient: httpClient}, nil
}

func (client *Client) ListByUsername(ctx context.Context, username string) ([]domain.Reservation, error) {
	var response []reservationResponse
	if err := client.do(ctx, http.MethodGet, "/api/v1/reservations", username, nil, http.StatusOK, &response); err != nil {
		return nil, mapError(err)
	}
	reservations := make([]domain.Reservation, 0, len(response))
	for _, item := range response {
		reservation, err := item.toDomain(username)
		if err != nil {
			return nil, err
		}
		reservations = append(reservations, reservation)
	}
	return reservations, nil
}

func (client *Client) CountByUsernameAndStatus(ctx context.Context, username string, status domain.ReservationStatus) (int, error) {
	var response countResponse
	path := "/api/v1/reservations/count?status=" + url.QueryEscape(string(status))
	if err := client.do(ctx, http.MethodGet, path, username, nil, http.StatusOK, &response); err != nil {
		return 0, mapError(err)
	}
	return response.Count, nil
}

func (client *Client) GetByUIDAndUsername(ctx context.Context, reservationUID uuid.UUID, username string) (domain.Reservation, error) {
	var response reservationResponse
	path := "/api/v1/reservations/" + reservationUID.String()
	if err := client.do(ctx, http.MethodGet, path, username, nil, http.StatusOK, &response); err != nil {
		return domain.Reservation{}, mapError(err)
	}
	return response.toDomain(username)
}

func (client *Client) Rent(ctx context.Context, username string, libraryUID, bookUID uuid.UUID, tillDate time.Time) (domain.Reservation, error) {
	var response reservationResponse
	body := rentRequest{BookUID: bookUID, LibraryUID: libraryUID, TillDate: tillDate.Format(dateLayout)}
	if err := client.do(ctx, http.MethodPost, "/api/v1/reservations", username, body, http.StatusCreated, &response); err != nil {
		return domain.Reservation{}, mapError(err)
	}
	return response.toDomain(username)
}

func (client *Client) Return(ctx context.Context, reservationUID uuid.UUID, username string, returnDate time.Time) (domain.Reservation, error) {
	var response reservationResponse
	path := "/api/v1/reservations/" + reservationUID.String() + "/return"
	if err := client.do(ctx, http.MethodPost, path, username, returnRequest{Date: returnDate.Format(dateLayout)}, http.StatusOK, &response); err != nil {
		return domain.Reservation{}, mapError(err)
	}
	return response.toDomain(username)
}

func (client *Client) do(ctx context.Context, method, path, username string, body any, expectedStatus int, result any) error {
	endpoint := *client.baseURL
	pathParts := strings.SplitN(path, "?", 2)
	endpoint.Path = strings.TrimRight(endpoint.Path, "/") + pathParts[0]
	endpoint.RawQuery = ""
	if len(pathParts) == 2 {
		endpoint.RawQuery = pathParts[1]
	}
	endpoint.Fragment = ""

	var requestBody io.Reader
	if body != nil {
		encoded, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("encode reservation request: %w", err)
		}
		requestBody = bytes.NewReader(encoded)
	}
	request, err := http.NewRequestWithContext(ctx, method, endpoint.String(), requestBody)
	if err != nil {
		return fmt.Errorf("create reservation request: %w", err)
	}
	request.Header.Set("X-User-Name", username)
	request.Header.Set("Accept", "application/json")
	if body != nil {
		request.Header.Set("Content-Type", "application/json")
	}
	response, err := client.httpClient.Do(request)
	if err != nil {
		return fmt.Errorf("call reservation service: %w", err)
	}
	defer response.Body.Close()
	limitedBody := io.LimitReader(response.Body, maxResponseBodySize)
	if response.StatusCode != expectedStatus {
		var serviceResponse errorResponse
		if err := json.NewDecoder(limitedBody).Decode(&serviceResponse); err != nil || serviceResponse.Message == "" {
			serviceResponse.Message = http.StatusText(response.StatusCode)
		}
		return &ServiceError{StatusCode: response.StatusCode, Message: serviceResponse.Message}
	}
	if err := json.NewDecoder(limitedBody).Decode(result); err != nil {
		return fmt.Errorf("decode reservation response: %w", err)
	}
	return nil
}

func mapError(err error) error {
	if serviceErr, ok := errors.AsType[*ServiceError](err); ok {
		switch serviceErr.StatusCode {
		case http.StatusNotFound:
			return domain.ErrReservationNotFound
		case http.StatusConflict:
			return domain.ErrReservationNotRented
		}
	}
	return err
}
