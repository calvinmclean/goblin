package dns

import (
	"bufio"
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

// Client is used to get IPs from the server over HTTP
type Client struct {
	addr string
}

func NewHTTPClient(addr string) (Client, error) {
	return Client{addr}, nil
}

func (c Client) GetIP(ctx context.Context, subdomain string) (string, error) {
	u := url.URL{
		Scheme: "http",
		Host:   c.addr,
		Path:   fmt.Sprintf("allocate/%s", subdomain),
	}

	resp, err := http.Post(u.String(), "", http.NoBody)
	if err != nil {
		return "", fmt.Errorf("failed to send request to server: %w", err)
	}

	ip, err := bufio.NewReader(resp.Body).ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("failed to read request body: %w", err)
	}
	ip = strings.TrimSpace(ip)

	if ip == "server error" {
		return "", fmt.Errorf("failed to get IP: %s", ip)
	}

	// keep the connection open until the context is done
	go func() {
		select {
		case <-ctx.Done():
			resp.Body.Close()
			return
		}
	}()

	return ip, nil
}

func (c Client) RegisterFallback(subdomain, address string) error {
	vals := url.Values{}
	vals.Add("address", address)
	u := url.URL{
		Scheme:   "http",
		Host:     c.addr,
		Path:     fmt.Sprintf("register/%s", subdomain),
		RawQuery: vals.Encode(),
	}

	resp, err := http.Post(u.String(), "", http.NoBody)
	if err != nil {
		return fmt.Errorf("failed to send request to server: %w", err)
	}

	if resp.StatusCode != http.StatusAccepted {
		return fmt.Errorf("unexpected response status: %d", resp.StatusCode)
	}

	return nil
}
