package wiseoldman

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
	"sync"
	"time"
)

const ApiUrl string = "https://api.wiseoldman.net/v2/players/"

type Client struct {
	baseUrl    *url.URL
	httpClient *http.Client
	limiter    *RateLimiter
}

type RateLimiter struct {
	mu          sync.Mutex
	interval    time.Duration
	nextRequest time.Time
	lastHeaders RateLimitHeaders
}

func NewRateLimiter(interval time.Duration) *RateLimiter {
	return &RateLimiter{interval: interval}
}

func (r *RateLimiter) Wait() {
	if r == nil || r.interval <= 0 {
		return
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	if now.Before(r.nextRequest) {
		time.Sleep(time.Until(r.nextRequest))
	}

	r.nextRequest = time.Now().Add(r.interval)
}

func (r *RateLimiter) Update(headers RateLimitHeaders) {
	if r == nil {
		return
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.lastHeaders = headers
}

func NewClient(apiUrl string) (*Client, error) {
	if apiUrl == "" {
		return nil, errors.New("apiUrl is empty")
	}

	parsedUrl, err := url.Parse(apiUrl)
	if err != nil {
		return nil, err
	}

	interval := time.Duration(0)
	if strings.Contains(parsedUrl.Host, "wiseoldman.net") {
		interval = time.Minute / 20
	}

	return &Client{
		baseUrl:    parsedUrl,
		httpClient: http.DefaultClient,
		limiter:    NewRateLimiter(interval),
	}, nil
}

func (c *Client) SendUpdateForPlayer(username string) (PlayerUpdate, error) {
	c.limiter.Wait()

	u := *c.baseUrl
	u.Path = path.Join(u.Path, username)

	response, err := c.httpClient.Post(u.String(), "application/json", nil)
	if err != nil {
		return PlayerUpdate{}, fmt.Errorf("error posting update for user %s: %w", username, err)
	}
	defer response.Body.Close()

	rateLimitHeaders := rateLimitHeadersFromResponse(response)
	c.limiter.Update(rateLimitHeaders)

	if err := checkApiResponse(response); err != nil {
		return PlayerUpdate{}, fmt.Errorf("error posting update for user %s: %w", username, err)
	}

	var apiSuccess PlayerUpdate
	if err := json.NewDecoder(response.Body).Decode(&apiSuccess); err != nil {
		return PlayerUpdate{}, fmt.Errorf("error posting update for user %s: %w", username, err)
	}

	apiSuccess.RateLimitHeaders = rateLimitHeaders

	return apiSuccess, nil
}

func (c *Client) GetPlayerData(userName string, queryParams *PlayerGainsParams) (PlayerGains, error) {
	c.limiter.Wait()

	if err := queryParams.Validate(); err != nil {
		return PlayerGains{}, err
	}

	u := *c.baseUrl
	u.Path = path.Join(u.Path, userName, "gained")
	u.RawQuery = queryParams.QueryValues().Encode()

	response, err := c.httpClient.Get(u.String())
	if err != nil {
		return PlayerGains{}, fmt.Errorf("Error getting player data for: %q:  %w", userName, err)
	}
	defer response.Body.Close()

	rateLimitHeaders := rateLimitHeadersFromResponse(response)
	c.limiter.Update(rateLimitHeaders)

	if err := checkApiResponse(response); err != nil {
		return PlayerGains{}, fmt.Errorf("Error getting player data for: %q:  %w", userName, err)
	}

	var apiSuccess PlayerGains
	if err := json.NewDecoder(response.Body).Decode(&apiSuccess); err != nil {
		return PlayerGains{}, fmt.Errorf("Error decoding player data for: %q:  %w", userName, err)
	}

	apiSuccess.PlayerName = userName
	apiSuccess.RateLimitHeaders = rateLimitHeaders

	return apiSuccess, nil
}

func rateLimitHeadersFromResponse(response *http.Response) RateLimitHeaders {
	return RateLimitHeaders{
		RateLimit:          headerInt(response, "RateLimit-Limit"),
		RateLimitRemaining: headerInt(response, "RateLimit-Remaining"),
		RateLimitReset:     headerInt(response, "RateLimit-Reset"),
	}
}

func headerInt(response *http.Response, name string) int {
	value, err := strconv.Atoi(response.Header.Get(name))
	if err != nil {
		return 0
	}

	return value
}

func checkApiResponse(response *http.Response) error {
	if response.StatusCode >= http.StatusOK && response.StatusCode < http.StatusMultipleChoices {
		return nil
	}

	var apiError ApiError
	if err := json.NewDecoder(response.Body).Decode(&apiError); err != nil {
		return err
	}

	return apiError
}
