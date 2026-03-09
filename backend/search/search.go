package search

import (
	"regexp"
	"sort"
	"strings"

	"gorm.io/gorm"
)

type SearchResult struct {
	URL   string  `json:"url"`
	Score float64 `json:"score"`
	Title string  `json:"title"`
	PR    float64 `json:"pr"`
}

type wordMatchRow struct {
	PageID uint
	URL    string
	Title  string
	PR     float64
	Word   string
	Count  uint
}

type pageRow struct {
	ID    uint
	URL   string
	Title string
	PR    float64
}

/*
 * SearchDataset executes a search against a specific dataset and returns ranked results.
 * The result score is the content score for the query, optionally boosted by PageRank.
 * If fewer than limit pages match the query, zero-score pages are appended until limit
 * results are returned, as required by the assignment.
 *
 * @param db       database connection
 * @param dataset  dataset identifier to search within
 * @param query    raw user query string
 * @param boost    whether to boost the content score using PageRank
 * @param limit    requested number of results
 * @return         ranked search results and an error if the search fails
 */
func SearchDataset(db *gorm.DB, dataset string, query string, boost bool, limit int) ([]SearchResult, error) {
	limit = clampLimit(limit)

	terms := tokenizeQuery(query)
	results := make([]SearchResult, 0, limit)
	seen := make(map[uint]bool)

	if len(terms) > 0 {
		rows, err := fetchMatchingWordCounts(db, dataset, terms)
		if err != nil {
			return nil, err
		}

		scored := make(map[uint]*SearchResult)

		// Add score to struct
		for _, row := range rows {
			entry, exists := scored[row.PageID]
			if !exists {
				title := row.Title
				if title == "" {
					title = row.URL
				}

				entry = &SearchResult{
					URL:   row.URL,
					Title: title,
					PR:    row.PR,
					Score: 0,
				}
				scored[row.PageID] = entry
			}

			entry.Score += float64(row.Count)
		}

		// Boost Results
		for pageID, entry := range scored {
			if boost {
				entry.Score *= entry.PR
			}
			results = append(results, *entry)
			seen[pageID] = true
		}

		// Sort in order of Score & PR
		sort.Slice(results, func(i, j int) bool {
			if results[i].Score == results[j].Score {
				if results[i].PR == results[j].PR {
					return results[i].URL < results[j].URL
				}
				return results[i].PR > results[j].PR
			}
			return results[i].Score > results[j].Score
		})
	}

	if len(results) < limit {
		fillers, err := fetchZeroScorePages(db, dataset, seen, limit-len(results))
		if err != nil {
			return nil, err
		}

		for _, page := range fillers {
			title := page.Title
			if title == "" {
				title = page.URL
			}

			results = append(results, SearchResult{
				URL:   page.URL,
				Title: title,
				PR:    page.PR,
				Score: 0,
			})
		}
	}

	if len(results) > limit {
		results = results[:limit]
	}

	return results, nil
}

/*
 * fetchMatchingWordCounts retrieves all word-count rows for query terms within a dataset.
 *
 * @param db       database connection
 * @param dataset  dataset identifier
 * @param terms    tokenized query terms
 * @return         matching joined rows and an error if the query fails
 */
func fetchMatchingWordCounts(db *gorm.DB, dataset string, terms []string) ([]wordMatchRow, error) {
	var rows []wordMatchRow

	err := db.
		Table("word_counts").
		Select("word_counts.page_id, pages.url, pages.title, pages.page_rank as pr, word_counts.word, word_counts.count").
		Joins("JOIN pages ON pages.id = word_counts.page_id").
		Where("pages.dataset = ?", dataset).
		Where("word_counts.word IN ?", terms).
		Find(&rows).Error

	if err != nil {
		return nil, err
	}

	return rows, nil
}

/*
 * fetchZeroScorePages retrieves pages from the dataset that are not already present in the
 * result set so that the caller can satisfy the required result limit even when scores are zero.
 * Fallback pages are ordered by PageRank descending, then by ID ascending.
 *
 * @param db         database connection
 * @param dataset    dataset identifier
 * @param excluded   page IDs already included in results
 * @param limit      maximum number of filler pages to return
 * @return           filler pages and an error if the query fails
 */
func fetchZeroScorePages(db *gorm.DB, dataset string, excluded map[uint]bool, limit int) ([]pageRow, error) {
	var pages []pageRow

	query := db.
		Table("pages").
		Select("id, url, title, page_rank as pr").
		Where("dataset = ?", dataset).
		Order("page_rank DESC").
		Order("id ASC")

	if len(excluded) > 0 {
		ids := make([]uint, 0, len(excluded))
		for id := range excluded {
			ids = append(ids, id)
		}
		query = query.Where("id NOT IN ?", ids)
	}

	if limit > 0 {
		query = query.Limit(limit)
	}

	if err := query.Find(&pages).Error; err != nil {
		return nil, err
	}

	return pages, nil
}

/*
 * clampLimit normalizes a requested result limit to the assignment bounds.
 *
 * @param limit requested result limit
 * @return      normalized limit in the range [1, 50]
 */
func clampLimit(limit int) int {
	if limit < 1 {
		return 10
	}
	if limit > 50 {
		return 50
	}
	return limit
}

var nonAlphaNum = regexp.MustCompile(`[^a-zA-Z0-9]+`)

/*
 * tokenizeQuery lowercases and tokenizes a raw query string into searchable terms.
 *
 * @param query raw user query
 * @return      normalized query terms
 */
func tokenizeQuery(query string) []string {
	query = strings.ToLower(strings.TrimSpace(query))
	if query == "" {
		return nil
	}

	parts := nonAlphaNum.Split(query, -1)
	seen := make(map[string]bool)
	terms := make([]string, 0, len(parts))

	for _, part := range parts {
		if part == "" {
			continue
		}
		if seen[part] {
			continue
		}
		seen[part] = true
		terms = append(terms, part)
	}

	return terms
}