package rating

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
	return fmt.Sprintf("rating service returned status %d: %s", err.StatusCode, err.Message)
}

func NewClient(rawBaseURL string, httpClient *http.Client) (*Client, error) {
	baseURL, err := url.Parse(strings.TrimSpace(rawBaseURL))
	if err != nil {
		return nil, fmt.Errorf("parse rating service URL: %w", err)
	}
	if baseURL.Scheme != "http" && baseURL.Scheme != "https" {
		return nil, fmt.Errorf("rating service URL must use http or https scheme")
	}
	if baseURL.Host == "" {
		return nil, fmt.Errorf("rating service URL must contain a host")
	}
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 5 * time.Second}
	}
	return &Client{baseURL: baseURL, httpClient: httpClient}, nil
}

func (client *Client) GetByUsername(ctx context.Context, username string) (domain.Rating, error) {
	var response ratingResponse
	if err := client.do(ctx, http.MethodGet, username, nil, http.StatusOK, &response); err != nil {
		return domain.Rating{}, mapError(err)
	}
	return domain.Rating{Username: username, StarsCount: response.Stars}, nil
}

func (client *Client) Create(ctx context.Context, rating domain.Rating) error {
	return mapError(client.do(ctx, http.MethodPost, rating.Username, setRatingRequest{Stars: rating.StarsCount}, http.StatusCreated, nil))
}

func (client *Client) UpdateRating(ctx context.Context, rating domain.Rating) (domain.Rating, error) {
	var response ratingResponse
	if err := client.do(ctx, http.MethodPut, rating.Username, setRatingRequest{Stars: rating.StarsCount}, http.StatusOK, &response); err != nil {
		return domain.Rating{}, mapError(err)
	}
	return domain.Rating{Username: rating.Username, StarsCount: response.Stars}, nil
}

func (client *Client) AddStars(ctx context.Context, username string, delta int) (domain.Rating, error) {
	var response ratingResponse
	if err := client.do(ctx, http.MethodPatch, username, addStarsRequest{Delta: delta}, http.StatusOK, &response); err != nil {
		return domain.Rating{}, mapError(err)
	}
	return domain.Rating{Username: username, StarsCount: response.Stars}, nil
}

func (client *Client) do(ctx context.Context, method, username string, body any, expectedStatus int, result any) error {
	endpoint := *client.baseURL
	endpoint.Path = strings.TrimRight(endpoint.Path, "/") + "/api/v1/rating"
	endpoint.RawQuery = ""
	endpoint.Fragment = ""

	var requestBody io.Reader
	if body != nil {
		encoded, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("encode rating request: %w", err)
		}
		requestBody = bytes.NewReader(encoded)
	}
	request, err := http.NewRequestWithContext(ctx, method, endpoint.String(), requestBody)
	if err != nil {
		return fmt.Errorf("create rating request: %w", err)
	}
	request.Header.Set("X-User-Name", username)
	request.Header.Set("Accept", "application/json")
	if body != nil {
		request.Header.Set("Content-Type", "application/json")
	}

	response, err := client.httpClient.Do(request)
	if err != nil {
		return fmt.Errorf("call rating service: %w", err)
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
	if result == nil {
		_, _ = io.Copy(io.Discard, limitedBody)
		return nil
	}
	if err := json.NewDecoder(limitedBody).Decode(result); err != nil {
		return fmt.Errorf("decode rating response: %w", err)
	}
	return nil
}

func mapError(err error) error {
	if serviceErr, ok := errors.AsType[*ServiceError](err); ok {
		switch serviceErr.StatusCode {
		case http.StatusNotFound:
			return domain.ErrRatingNotFound
		case http.StatusConflict:
			return domain.ErrRatingAlreadyExists
		}
	}
	return err
}
