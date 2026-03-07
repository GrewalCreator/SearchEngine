package persistence

import (
	"errors"
	"gorm.io/gorm"
)

func GetPage(db *gorm.DB, url string, dataset string) (*Page, error) {
	var page Page

	result := db.First(&page, "url = ? AND dataset = ?", url, dataset)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}

		return nil, result.Error
	}

	return &page, nil
}

// CreatePage creates a new page entry in the database
func CreatePage(db *gorm.DB, url string, dataset string) (*Page, error) {
	if url == "" {
		return nil, errors.New("url cannot be empty")
	}

	if dataset == "" {
		return nil, errors.New("dataset cannot be empty")
	}

	page := Page{
		URL:     url,
		Dataset: dataset,
	}

	result := db.Create(&page)
	if result.Error != nil {
		return nil, result.Error
	}

	return &page, nil
}

/*
* GetOrCreatePage Gets the page at url
* If the page does not exist, it will create it and return the page
*/
func GetOrCreatePage(db *gorm.DB, url string, dataset string) (*Page, error) {
	if url == "" {
		return nil, errors.New("url cannot be empty")
	}

	if dataset == "" {
		return nil, errors.New("dataset cannot be empty")
	}

	page := Page{
		URL:     url,
		Dataset: dataset,
	}

	result := db.Where("url = ? AND dataset = ?", url, dataset).FirstOrCreate(&page)
	if result.Error != nil {
		return nil, result.Error
	}

	return &page, nil
}

// UpdatePageTitle updates the title of page with ID pageID
func UpdatePageTitle(db *gorm.DB, pageID uint, title string) error {
	result := db.Model(&Page{}).
		Where("id = ?", pageID).
		Update("title", title)

	return result.Error
}

// UpdatePageRank updates the pageRank of page with ID pageID
func UpdatePageRank(db *gorm.DB, pageID uint, pageRank float64) error {
	result := db.Model(&Page{}).
		Where("id = ?", pageID).
		Update("page_rank", pageRank)

	return result.Error
}

// UpdatePage updates multiple values of page record
func UpdatePage(db *gorm.DB, pageID uint, updates map[string]any) error {
	result := db.Model(&Page{}).
		Where("id = ?", pageID).
		Updates(updates)

	return result.Error
}

// DeletePage safely deletes a page and removes related links and word counts
func DeletePage(db *gorm.DB, pageID uint) error {
	tx := db.Begin()

	if tx.Error != nil {
		return tx.Error
	}

	if err := tx.Where("from_page_id = ? OR to_page_id = ?", pageID, pageID).Delete(&Link{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Where("page_id = ?", pageID).Delete(&WordCount{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Delete(&Page{}, pageID).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

// CreateLink creates a link from page with Id fromID to page with ID toID
func CreateLink(db *gorm.DB, fromID uint, toID uint) error {
	link := Link{
		FromPageID: fromID,
		ToPageID:   toID,
	}

	result := db.Where("from_page_id = ? AND to_page_id = ?", fromID, toID).
		FirstOrCreate(&link)

	return result.Error
}

// Returns all outgoing links of page with ID pageID
func GetOutgoingLinks(db *gorm.DB, pageID uint) ([]Link, error) {
	var links []Link

	result := db.Where("from_page_id = ?", pageID).Find(&links)
	if result.Error != nil {
		return nil, result.Error
	}

	return links, nil
}

// Returns all incoming links of page with ID pageID
func GetIncomingLinks(db *gorm.DB, pageID uint) ([]Link, error) {
	var links []Link

	result := db.Where("to_page_id = ?", pageID).Find(&links)
	if result.Error != nil {
		return nil, result.Error
	}

	return links, nil
}

// Saves/Updates the word count map of page with ID pageID
func SaveWordCounts(db *gorm.DB, pageID uint, counts map[string]uint) error {
	tx := db.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	if err := tx.Where("page_id = ?", pageID).Delete(&WordCount{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	for word, count := range counts {
		wordCount := WordCount{
			PageID: pageID,
			Word:   word,
			Count:  count,
		}

		if err := tx.Create(&wordCount).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit().Error
}