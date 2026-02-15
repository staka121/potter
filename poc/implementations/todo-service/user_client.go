package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// UserClient handles communication with user-service
type UserClient struct {
	baseURL    string
	httpClient *http.Client
}

// ValidateUserRequest is the request body for user validation
type ValidateUserRequest struct {
	UserID string `json:"user_id"`
}

// ValidateUserResponse is the response from user validation
type ValidateUserResponse struct {
	Valid bool        `json:"valid"`
	User  interface{} `json:"user,omitempty"`
}

// NewUserClient creates a new user service client
func NewUserClient(baseURL string) *UserClient {
	return &UserClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

// ValidateUser validates if a user exists by calling user-service
func (c *UserClient) ValidateUser(userID string) (bool, error) {
	url := fmt.Sprintf("%s/api/v1/users/validate", c.baseURL)

	reqBody := ValidateUserRequest{
		UserID: userID,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return false, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(jsonData))
	if err != nil {
		return false, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return false, fmt.Errorf("failed to call user-service: %w", err)
	}
	defer resp.Body.Close()

	// If the response is 404, the user is not valid
	if resp.StatusCode == http.StatusNotFound {
		return false, nil
	}

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("unexpected status code from user-service: %d", resp.StatusCode)
	}

	var validateResp ValidateUserResponse
	if err := json.NewDecoder(resp.Body).Decode(&validateResp); err != nil {
		return false, fmt.Errorf("failed to decode response: %w", err)
	}

	return validateResp.Valid, nil
}
