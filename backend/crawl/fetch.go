package crawl

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

func FetchPage(pageURL string) ([]byte, error) {

	// Setup Http Client with a 10 Second timeout
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// Get the page at pageURL
	resp, err := client.Get(pageURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch %s: %w", pageURL, err)
	}

	// Wait until just before function completes to Close() the body stream
	// Must be placed right after Open or defer code may never get executed (open connection done in Get())
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code %d for %s", resp.StatusCode, pageURL)
	}

	// If contentType exists and its not of type text/html, return error (we do not handle non html pages)
	contentType := resp.Header.Get("Content-Type")
	if contentType != "" && !strings.Contains(strings.ToLower(contentType), "text/html") {
		return nil, fmt.Errorf("non-html content at %s: %s", pageURL, contentType)
	}

	// Reads response body stream as returns a byte array or error
	// resp.Body acts more as a file so it must be read into a usable format
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body for %s: %w", pageURL, err)
	}

	return body, nil
}

