package dns

import (
	"bufio"
	"context"
	"fmt"
	"io"
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

	if resp.StatusCode != http.StatusCreated {
		fmt.Println("STTUS", resp.StatusCode)
		printResponseBody(resp)
		return "", fmt.Errorf("unexpected response status: %d", resp.StatusCode)
	}

	ip, err := bufio.NewReader(resp.Body).ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("failed to read request body: %w", err)
	}
	ip = strings.TrimSpace(ip)

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
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		printResponseBody(resp)
		return fmt.Errorf("unexpected response status: %d", resp.StatusCode)
	}

	return nil
}

func printResponseBody(r *http.Response) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		fmt.Printf("error reading response body: %v\n", err)
	}
	fmt.Println(string(body))
}
