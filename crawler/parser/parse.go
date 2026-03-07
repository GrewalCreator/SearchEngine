package parser

import (
	"bytes"
	"net/url"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type ParsedPage struct {
	URL        string
	Title      string
	Links      []string
	WordCounts map[string]uint
}

func ParsePage(pageURL string, body []byte) (*ParsedPage, error) {

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	baseURL, err := url.Parse(pageURL)
	if err != nil {
		return nil, err
	}

	result := &ParsedPage{
		URL:        pageURL,
		Title:      "",
		Links:      []string{},
		WordCounts: make(map[string]uint),
	}

	// Extract the title value
	result.Title = strings.TrimSpace(doc.Find("title").First().Text())

	// Extract all links
	seen := make(map[string]bool)

	doc.Find("a").Each(func(i int, s *goquery.Selection) {

		href, exists := s.Attr("href")
		if !exists {
			return
		}

		// Clean link (especially for personal page)
		link := normalizeLink(baseURL, href)
		if link == "" {
			return
		}

		// Filter duplicates then add to struct
		if !seen[link] {
			seen[link] = true
			result.Links = append(result.Links, link)
		}
	})

	// Extract text and count occurences
	text := doc.Text()
	addWords(result.WordCounts, text)

	return result, nil
}


// Clean links 
func normalizeLink(base *url.URL, raw string) string {

	raw = strings.TrimSpace(raw)
	if raw == "" || strings.HasPrefix(raw, "#") {
		return ""
	}

	u, err := url.Parse(raw)
	if err != nil {
		return ""
	}

	resolved := base.ResolveReference(u)
	resolved.Fragment = ""

	if resolved.Scheme != "http" && resolved.Scheme != "https" {
		return ""
	}

	return resolved.String()
}


var nonLetter = regexp.MustCompile(`[^a-zA-Z0-9]+`)


// Count occurences for each word in the text
func addWords(counts map[string]uint, text string) {

	text = strings.ToLower(text)

	words := nonLetter.Split(text, -1)

	for _, w := range words {
		if w == "" {
			continue
		}
		counts[w]++
	}
}