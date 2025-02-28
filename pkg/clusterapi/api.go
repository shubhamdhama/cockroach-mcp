package clusterapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

type APIClient struct {
	BaseURL       string
	HTTPClient    *http.Client
	SessionCookie *http.Cookie
}

var apiClient *APIClient

func GetClient() *APIClient {
	return apiClient
}

func InitAPIClient() {
	baseURL := "http://localhost:8080"
	if urlEnv := os.Getenv("COCKROACH_API_URL"); urlEnv != "" {
		baseURL = urlEnv
	}
	client := &APIClient{
		BaseURL: baseURL,
		HTTPClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
	if err := client.login(); err != nil {
		log.Fatalf("Failed to login for cockroach API: %v", err)
	}
	apiClient = client
}

func (c *APIClient) login() error {
	username := os.Getenv("COCKROACH_API_USERNAME")
	if username == "" {
		log.Fatal("Failed to get username")
	}
	password := os.Getenv("COCKROACH_API_PASSWORD")
	if password == "" {
		log.Fatal("Failed to get password")
	}
	loginURL := fmt.Sprintf("%s/login", c.BaseURL)
	payload := map[string]string{
		"username": username,
		"password": password,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", loginURL, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("login failed: %s", string(bodyBytes))
	}
	var sessionCookie *http.Cookie
	cookies := resp.Cookies()
	for _, cookie := range cookies {
		if cookie.Name == "session" {
			sessionCookie = cookie
		}
	}
	if sessionCookie == nil {
		return fmt.Errorf("no session cookie received")
	}
	c.SessionCookie = sessionCookie
	log.Print("Login to cockroach cluster APIs successful.")
	log.Print(sessionCookie)
	return nil
}

func (c *APIClient) QueryTimeseries(
	ctx context.Context, tenant string, startNanos, endNanos int64, queryName string,
) (*TimeseriesQueryResponse, error) {
	queryURL := fmt.Sprintf("%s/ts/query", c.BaseURL)
	reqBody := TimeseriesQueryRequest{
		StartNanos: startNanos,
		EndNanos:   endNanos,
		Queries: []Query{
			{
				Name: queryName,
			},
		},
	}
	payloadBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}
	log.Printf("payload: %s", string(payloadBytes))
	req, err := http.NewRequestWithContext(ctx, "POST", queryURL, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return nil, err
	}
	responseBody, err := c.sendRequest(req, tenant)
	if err != nil {
		return nil, err
	}
	var response TimeseriesQueryResponse
	if err := json.Unmarshal(responseBody, &response); err != nil {
		log.Printf("Failed to unmarshal response: %v, %s", err, string(responseBody))
		return nil, err
	}
	return &response, nil
}

func (c *APIClient) sendRequest(req *http.Request, tenant string) ([]byte, error) {
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(c.SessionCookie)
	req.AddCookie(&http.Cookie{Name: "tenant", Value: tenant})
	var responseBody []byte
	for attempt := range 2 {
		resp, err := c.HTTPClient.Do(req)
		if err != nil {
			log.Printf("Request failed: %v", err)
			return nil, err
		}
		defer resp.Body.Close()
		if resp.StatusCode == http.StatusOK {
			if responseBody, err = io.ReadAll(resp.Body); err != nil {
				log.Printf("Failed to read response: %v", err)
				return nil, err
			}
			break
		}
		if attempt == 0 && resp.StatusCode == http.StatusUnauthorized {
			log.Print("Unauthorized request, retrying after login")
			c.login()
			continue
		}
		bodyBytes, _ := io.ReadAll(resp.Body)
		log.Printf("Received failed response: %s", bodyBytes)
		return nil, fmt.Errorf("query failed: %s, %v", http.StatusText(resp.StatusCode), resp)
	}
	return responseBody, nil
}
