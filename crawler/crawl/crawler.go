package crawl

import (
	"log"
	"net/url"
	"strings"

	"crawler/parser"
	"crawler/persistence"

	"gorm.io/gorm"
)

type Crawler struct {
	DB            *gorm.DB
	Dataset       string
	StartURL      string
	Limit         int
	AllowedPrefix []string

	visited map[string]bool
	queued  map[string]bool
	queue   []string
}

/*
 * NewCrawler constructs a new sequential crawler instance.
 *
 * @param db        database connection used for persistence
 * @param startURL  seed URL where the crawl begins
 * @param dataset   dataset identifier (e.g. fruitsA)
 * @param limit     maximum number of pages to crawl
 * @param _         unused worker count parameter kept for API compatibility
 * @return          initialized crawler instance
 */
func NewCrawler(db *gorm.DB, startURL string, dataset string, limit int, _ int) *Crawler {
	return &Crawler{
		DB:            db,
		Dataset:       dataset,
		StartURL:      startURL,
		Limit:         limit,
		AllowedPrefix: buildAllowedPrefixes(startURL, dataset),
		visited:       make(map[string]bool),
		queued:        map[string]bool{startURL: true},
		queue:         []string{startURL},
	}
}

/*
 * Crawl starts the sequential crawl from the seed URL and continues until
 * the queue is empty or the crawl limit is reached.
 *
 * @return error if a fatal crawl error occurs
 */
func (c *Crawler) Crawl() error {
	crawledCount := 0

	for len(c.queue) > 0 && crawledCount < c.Limit {
		pageURL := c.pop()

		if c.visited[pageURL] {
			continue
		}
		c.visited[pageURL] = true

		if !c.allowedURL(pageURL) {
			continue
		}

		log.Printf("crawling %s", pageURL)

		links, err := c.processPage(pageURL)
		if err != nil {
			log.Printf("failed %s: %v", pageURL, err)
			continue
		}

		crawledCount++

		for _, link := range links {
			if crawledCount >= c.Limit {
				break
			}

			if !c.allowedURL(link) {
				continue
			}

			if c.visited[link] || c.queued[link] {
				continue
			}

			c.queued[link] = true
			c.queue = append(c.queue, link)
		}
	}

	log.Printf("crawl complete: %d pages crawled", crawledCount)
	return nil
}

/*
 * processPage fetches and parses a page, persists its metadata,
 * stores word counts, and records discovered links.
 *
 * @param pageURL URL of the page to process
 * @return slice of discovered links and an error if processing fails
 */
func (c *Crawler) processPage(pageURL string) ([]string, error) {
	body, err := FetchPage(pageURL)
	if err != nil {
		return nil, err
	}

	parsed, err := parser.ParsePage(pageURL, body)
	if err != nil {
		return nil, err
	}

	page, err := persistence.GetOrCreatePage(c.DB, parsed.URL, c.Dataset)
	if err != nil {
		return nil, err
	}

	if parsed.Title != "" {
		if err := persistence.UpdatePageTitle(c.DB, page.ID, parsed.Title); err != nil {
			return nil, err
		}
	}

	if err := persistence.SaveWordCounts(c.DB, page.ID, parsed.WordCounts); err != nil {
		return nil, err
	}

	for _, linkURL := range parsed.Links {
		if !c.allowedURL(linkURL) {
			continue
		}

		linkedPage, err := persistence.GetOrCreatePage(c.DB, linkURL, c.Dataset)
		if err != nil {
			return nil, err
		}

		if err := persistence.CreateLink(c.DB, page.ID, linkedPage.ID); err != nil {
			return nil, err
		}
	}

	return parsed.Links, nil
}

/*
 * pop removes and returns the next URL from the crawl queue.
 *
 * @return next URL in FIFO order
 */
func (c *Crawler) pop() string {
	pageURL := c.queue[0]
	c.queue = c.queue[1:]
	return pageURL
}

/*
 * allowedURL verifies that a URL belongs to the permitted dataset prefixes.
 *
 * @param raw URL string to validate
 * @return true if the URL belongs to the allowed dataset
 */
func (c *Crawler) allowedURL(raw string) bool {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return false
	}

	for _, prefix := range c.AllowedPrefix {
		if strings.HasPrefix(raw, prefix) {
			return true
		}
	}

	return false
}

/*
 * buildAllowedPrefixes constructs the list of URL prefixes that are allowed
 * to be crawled for a given dataset.
 *
 * @param startURL initial seed URL
 * @param dataset  dataset identifier
 * @return slice of allowed URL prefixes
 */
func buildAllowedPrefixes(startURL string, dataset string) []string {
	u, err := url.Parse(startURL)
	if err != nil {
		return []string{startURL}
	}

	base := u.Scheme + "://" + u.Host

	if dataset == "fruitsA" {
		return []string{
			base + "/~avamckenney/fruitsA/",
			base + "/~avamckenney/fruitsB/",
		}
	}

	path := u.Path
	lastSlash := strings.LastIndex(path, "/")
	if lastSlash != -1 {
		return []string{base + path[:lastSlash+1]}
	}

	return []string{base + "/"}
}
