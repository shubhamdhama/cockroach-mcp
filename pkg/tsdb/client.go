package tsdb

import (
	"context"
	"fmt"
	"time"

	"github.com/sony/gobreaker"
	resty "resty.dev/v3"
)

const (
	tsQueryURL = "http://127.0.0.1:8080/ts/query"
)

var tsdbClient *TSDBClient

func init() {
	tsdbClient = NewTSDBClient()
}

type TSDBClient struct {
	client         *resty.Client
	circuitBreaker *gobreaker.CircuitBreaker
}

func NewTSDBClient() *TSDBClient {
	client := resty.New().
		SetRetryCount(3).
		SetRetryWaitTime(2 * time.Second).
		SetRetryMaxWaitTime(10 * time.Second).
		// SetBasicAuth(username, password).
		SetTimeout(10 * time.Second)

	cbSettings := gobreaker.Settings{
		Name:        "Query TSDB",
		MaxRequests: 5,
		Interval:    60 * time.Second,
		Timeout:     30 * time.Second,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			return counts.ConsecutiveFailures > 3
		},
	}

	circuitBreaker := gobreaker.NewCircuitBreaker(cbSettings)

	return &TSDBClient{
		client:         client,
		circuitBreaker: circuitBreaker,
	}
}

func (rc *TSDBClient) Query(ctx context.Context, req *TSQueryRequest) (*TSQueryResponse, error) {
	resp := TSQueryResponse{}
	result, err := rc.circuitBreaker.Execute(func() (interface{}, error) {
		resp, err := rc.client.R().
			SetBody(req).
			SetResult(&resp).
			Post(tsQueryURL)
		if err != nil {
			return nil, err
		}
		if resp.IsError() {
			return nil, fmt.Errorf("error: %v", resp.Status())
		}
		return resp.Result().(*TSQueryResponse), nil
	})

	if err != nil {
		return nil, err
	}

	return result.(*TSQueryResponse), nil
}

func Client() *TSDBClient {
	return tsdbClient
}
