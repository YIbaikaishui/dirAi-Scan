package middleware

import (
	"net/http"
	"net/url"
	"time"

	"golang.org/x/time/rate"
)

type RateLimiter struct {
	limiter *rate.Limiter
}

func NewRateLimiter(requestsPerSecond float64) *RateLimiter {
	return &RateLimiter{
		limiter: rate.NewLimiter(rate.Limit(requestsPerSecond), 1),
	}
}

func (r *RateLimiter) WrapClient(client *http.Client) *http.Client {
	transport := client.Transport
	if transport == nil {
		transport = http.DefaultTransport
	}

	client.Transport = &rateLimitedTransport{
		base:    transport,
		limiter: r.limiter,
	}

	return client
}

type rateLimitedTransport struct {
	base    http.RoundTripper
	limiter *rate.Limiter
}

func (t *rateLimitedTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	err := t.limiter.Wait(req.Context())
	if err != nil {
		return nil, err
	}
	return t.base.RoundTrip(req)
}

type ProxyClient struct {
	proxyURL *url.URL
}

func NewProxyClient(proxyAddr string) (*ProxyClient, error) {
	proxyURL, err := url.Parse(proxyAddr)
	if err != nil {
		return nil, err
	}

	return &ProxyClient{proxyURL: proxyURL},
		nil
}

func (p *ProxyClient) WrapClient(client *http.Client) *http.Client {
	transport := client.Transport
	if transport == nil {
		transport = http.DefaultTransport
	}

	if httpTransport, ok := transport.(*http.Transport); ok {
		httpTransport.Proxy = http.ProxyURL(p.proxyURL)
	}

	return client
}

type RetryClient struct {
	maxRetries  int
	retryDelay  time.Duration
	retryStatus []int
}

func NewRetryClient(maxRetries int, retryDelay time.Duration, retryStatus []int) *RetryClient {
	return &RetryClient{
		maxRetries:  maxRetries,
		retryDelay:  retryDelay,
		retryStatus: retryStatus,
	}
}

func (r *RetryClient) WrapClient(client *http.Client) *http.Client {
	transport := client.Transport
	if transport == nil {
		transport = http.DefaultTransport
	}

	client.Transport = &retryTransport{
		base:        transport,
		maxRetries:  r.maxRetries,
		retryDelay:  r.retryDelay,
		retryStatus: r.retryStatus,
	}

	return client
}

type retryTransport struct {
	base        http.RoundTripper
	maxRetries  int
	retryDelay  time.Duration
	retryStatus []int
}

func (t *retryTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	var resp *http.Response
	var err error

	for i := 0; i <= t.maxRetries; i++ {
		resp, err = t.base.RoundTrip(req)
		if err == nil && !t.shouldRetry(resp.StatusCode) {
			break
		}

		if i < t.maxRetries {
			time.Sleep(t.retryDelay)
		}
	}

	return resp, err
}

func (t *retryTransport) shouldRetry(statusCode int) bool {
	for _, code := range t.retryStatus {
		if code == statusCode {
			return true
		}
	}
	return false
}