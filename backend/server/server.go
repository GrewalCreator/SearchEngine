package server

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"

	"crawler/crawl"
	d_graph "crawler/d-graph"
	"crawler/persistence"
	"crawler/search"
	"crawler/util"

	"gorm.io/gorm"
)

type Server struct {
	db     *gorm.DB
	cfg    *util.Config
	router *http.ServeMux
}

/*
 * SearchResponse represents the JSON response for search endpoints
 */
type SearchResponse struct {
	Result []search.SearchResult `json:"result"`
}

/*
 * PageDetailsResponse represents the JSON response for the page details endpoint
 */
type PageDetailsResponse struct {
	URL           string          `json:"url"`
	Title         string          `json:"title"`
	Dataset       string          `json:"dataset"`
	PageRank      float64         `json:"pr"`
	IncomingLinks []string        `json:"incoming_links"`
	OutgoingLinks []string        `json:"outgoing_links"`
	WordCounts    map[string]uint `json:"word_counts"`
}

/*
 * pageDetailsRow is used internally to join page details queries
 */
type pageDetailsRow struct {
	ID       uint
	URL      string
	Title    string
	Dataset  string
	PageRank float64 `gorm:"column:page_rank"`
}

/*
 * wordCountRow is used internally for word count queries
 */
type wordCountRow struct {
	Word  string
	Count uint
}

/*
 * NewServer constructs a new REST server, initializes the database,
 * crawls configured datasets, computes PageRank, and registers routes
 *
 * @return initialized Server instance
 */
func NewServer() (*Server, error) {
	cfg, err := util.LoadConfig("../config.yml")
	if err != nil {
		return nil, err
	}

	db, err := persistence.InitalizeDatabase(cfg.SearchEngine.Database.StoragePath)
	if err != nil {
		return nil, err
	}

	s := &Server{
		db:     db,
		cfg:    cfg,
		router: http.NewServeMux(),
	}

	if err := s.initializeData(); err != nil {
		return nil, err
	}

	s.registerRoutes()

	return s, nil
}

/*
 * Start begins listening for HTTP requests on port 3000
 *
 * @return error if the server fails to start
 */
func (s *Server) Start() error {
	log.Println("REST server listening on :3000")
	return http.ListenAndServe(":3000", s.router)
}

/*
 * initializeData crawls configured datasets and computes PageRank
 *
 * @return error if crawling or ranking fails
 */
func (s *Server) initializeData() error {
	log.Println("database initialized")

	fruitsCrawler := crawl.NewCrawler(
		s.db,
		s.cfg.SearchEngine.Crawler.StartURL,
		"fruitsA",
		s.cfg.SearchEngine.Crawler.CrawlLimit,
		1,
	)

	if err := fruitsCrawler.Crawl(); err != nil {
		return err
	}

	log.Println("fruits crawl complete")

	if err := d_graph.ComputePageRank(s.db, "fruitsA", s.cfg.SearchEngine.Crawler.Damping, s.cfg.SearchEngine.Crawler.Iterations); err != nil {
		return err
	}

	log.Println("fruits PageRank complete")

	if s.cfg.SearchEngine.Crawler.CustomURL != "" {
		personalCrawler := crawl.NewCrawler(
			s.db,
			s.cfg.SearchEngine.Crawler.CustomURL,
			"personal",
			s.cfg.SearchEngine.Crawler.CrawlLimit,
			1,
		)

		if err := personalCrawler.Crawl(); err != nil {
			return err
		}

		log.Println("personal crawl complete")

		if err := d_graph.ComputePageRank(s.db, "personal", 0.85, 30); err != nil {
			return err
		}

		log.Println("personal PageRank complete")
	}

	return nil
}

/*
 * registerRoutes registers all REST endpoints.
 */
func (s *Server) registerRoutes() {
	s.router.HandleFunc("/info", s.infoHandler)
	s.router.HandleFunc("/health", s.healthHandler)
	s.router.HandleFunc("/fruitsA", s.fruitsHandler)
	s.router.HandleFunc("/personal", s.personalHandler)
	s.router.HandleFunc("/page", s.pageDetailsHandler)

	// Serve the frontend from ../frontend
	s.router.Handle("/", http.FileServer(http.Dir("../frontend")))
}

/*
 * infoHandler returns server registration info
 */
func (s *Server) infoHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{
		"name": "SearchEngine",
	})
}

/*
 * healthHandler returns a simple health response to check server status
 */
func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"status": "ok",
	})
}

/*
 * fruitsHandler executes a search against the fruits dataset
 */
func (s *Server) fruitsHandler(w http.ResponseWriter, r *http.Request) {
	s.handleSearch(w, r, "fruitsA")
}

/*
 * personalHandler executes a search against the personal dataset.
 */
func (s *Server) personalHandler(w http.ResponseWriter, r *http.Request) {
	s.handleSearch(w, r, "personal")
}

/*
 * handleSearch parses query parameters and returns ranked search results
 *
 * Supported query params:
 * q      - query string
 * boost  - true/false
 * limit  - 1..50
 */
func (s *Server) handleSearch(w http.ResponseWriter, r *http.Request, dataset string) {
	query := r.URL.Query().Get("q")

	boost := false
	if r.URL.Query().Get("boost") == "true" {
		boost = true
	}

	limit := 10
	if rawLimit := r.URL.Query().Get("limit"); rawLimit != "" {
		parsedLimit, err := strconv.Atoi(rawLimit)
		if err == nil {
			limit = parsedLimit
		}
	}

	results, err := search.SearchDataset(s.db, dataset, query, boost, limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	writeJSON(w, http.StatusOK, SearchResponse{
		Result: results,
	})
}

/*
 * pageDetailsHandler returns stored details for a specific page
 *
 * Required query params:
 * dataset - fruitsA or personal
 * url     - full page URL
 */
func (s *Server) pageDetailsHandler(w http.ResponseWriter, r *http.Request) {
	dataset := r.URL.Query().Get("dataset")
	pageURL := r.URL.Query().Get("url")

	if dataset == "" || pageURL == "" {
		writeError(w, http.StatusBadRequest, errors.New("dataset and url are required"))
		return
	}

	page, err := getPageByURLAndDataset(s.db, pageURL, dataset)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	if page == nil {
		writeError(w, http.StatusNotFound, errors.New("page not found"))
		return
	}

	outgoingLinks, err := getOutgoingLinkURLs(s.db, page.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	incomingLinks, err := getIncomingLinkURLs(s.db, page.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	wordCounts, err := getWordCounts(s.db, page.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	writeJSON(w, http.StatusOK, PageDetailsResponse{
		URL:           page.URL,
		Title:         page.Title,
		Dataset:       page.Dataset,
		PageRank:      page.PageRank,
		IncomingLinks: incomingLinks,
		OutgoingLinks: outgoingLinks,
		WordCounts:    wordCounts,
	})
}

/*
 * getPageByURLAndDataset retrieves a single page by URL and dataset
 *
 * @param db      database connection
 * @param pageURL page URL
 * @param dataset dataset identifier
 * @return matching page or nil if not found
 */
func getPageByURLAndDataset(db *gorm.DB, pageURL string, dataset string) (*pageDetailsRow, error) {
	var page pageDetailsRow

	result := db.
		Table("pages").
		Select("id, url, title, dataset, page_rank").
		Where("url = ? AND dataset = ?", pageURL, dataset).
		First(&page)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, result.Error
	}

	return &page, nil
}

/*
 * getOutgoingLinkURLs retrieves all outgoing link URLs for a page
 *
 * @param db     database connection
 * @param pageID page identifier
 * @return slice of outgoing linked page URLs
 */
func getOutgoingLinkURLs(db *gorm.DB, pageID uint) ([]string, error) {
	type row struct {
		URL string
	}

	var rows []row

	err := db.
		Table("links").
		Select("pages.url").
		Joins("JOIN pages ON pages.id = links.to_page_id").
		Where("links.from_page_id = ?", pageID).
		Find(&rows).Error
	if err != nil {
		return nil, err
	}

	urls := make([]string, 0, len(rows))
	for _, r := range rows {
		urls = append(urls, r.URL)
	}

	return urls, nil
}

/*
 * getIncomingLinkURLs retrieves all incoming link URLs for a page
 *
 * @param db     database connection
 * @param pageID page identifier
 * @return slice of incoming linked page URLs
 */
func getIncomingLinkURLs(db *gorm.DB, pageID uint) ([]string, error) {
	type row struct {
		URL string
	}

	var rows []row

	err := db.
		Table("links").
		Select("pages.url").
		Joins("JOIN pages ON pages.id = links.from_page_id").
		Where("links.to_page_id = ?", pageID).
		Find(&rows).Error
	if err != nil {
		return nil, err
	}

	urls := make([]string, 0, len(rows))
	for _, r := range rows {
		urls = append(urls, r.URL)
	}

	return urls, nil
}

/*
 * getWordCounts retrieves all stored word frequencies for a page
 *
 * @param db     database connection
 * @param pageID page identifier
 * @return map of word to frequency
 */
func getWordCounts(db *gorm.DB, pageID uint) (map[string]uint, error) {
	var rows []wordCountRow

	err := db.
		Table("word_counts").
		Select("word, count").
		Where("page_id = ?", pageID).
		Find(&rows).Error
	if err != nil {
		return nil, err
	}

	counts := make(map[string]uint, len(rows))
	for _, row := range rows {
		counts[row.Word] = row.Count
	}

	return counts, nil
}

/*
 * writeJSON writes a JSON response with the specified status code
 *
 * @param w      response writer
 * @param status HTTP status code
 * @param v      response payload
 */
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

/*
 * writeError writes a JSON error response
 *
 * @param w      response writer
 * @param status HTTP status code
 * @param err    error to serialize
 */
func writeError(w http.ResponseWriter, status int, err error) {
	writeJSON(w, status, map[string]any{
		"error": err.Error(),
	})
}