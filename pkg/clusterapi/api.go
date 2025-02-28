package clusterapi

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/shubhamdhama/cockroach-mcp/pkg/utils"
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
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true, // WARNING: Only use this for development!
				},
			},
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
	loginURL, err := url.JoinPath(c.BaseURL, "login")
	if err != nil {
		log.Fatalf("Failed to build login URL: %v", err)
	}

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
	return nil
}

func (c *APIClient) QueryTimeseries(
	ctx context.Context, tenant string, startNanos, endNanos int64, queryName string,
) (string, error) {
	queryURL, err := url.JoinPath(c.BaseURL, "ts", "query")
	if err != nil {
		return "", err
	}
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
		return "", err
	}
	log.Printf("payload: %s", string(payloadBytes))
	req, err := http.NewRequestWithContext(ctx, "POST", queryURL, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return "", err
	}
	responseBody, err := c.sendRequest(req, tenant)
	if err != nil {
		return "", err
	}
	var response TimeseriesQueryResponse
	if err := json.Unmarshal(responseBody, &response); err != nil {
		log.Printf("Failed to unmarshal response: %v, %s", err, string(responseBody))
		return "", err
	}

	cols := []string{"timestamp", "value"}
	var builder strings.Builder
	for _, result := range response.Results {
		var rows [][]any
		for _, queryResult := range result.Datapoints {
			nanos, err := strconv.ParseInt(queryResult.TimestampNanos, 10, 64)
			if err != nil {
				return "", err
			}
			t := time.Unix(0, nanos)
			formatted := t.Format(time.DateTime)
			rows = append(rows, []any{formatted, queryResult.Value})
		}

		fmt.Fprintf(&builder, "Result for query: %s\n%s\n",
			result.Query.Name, utils.FormatAsMarkdown(cols, rows))
	}

	return builder.String(), nil
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
