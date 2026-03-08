package persistence

import (
	"fmt"
	"os"
	"path/filepath"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

/*
- InitalizeDatabase inits the database using gorm
- @param  dbPath: Path to data storage
- @return *gorm.DB: Pointer to Database object
- @return error occuring during initalization
*/
func InitalizeDatabase(dbPath string) (*gorm.DB, error) {
	if dbPath == "" {
		return nil, fmt.Errorf("database path cannot be empty")
	}

	absPath, err := filepath.Abs(dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve database path: %w", err)
	}

	parentDir := filepath.Dir(absPath)
	if err := os.MkdirAll(parentDir, 0o755); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}

	db, err := gorm.Open(sqlite.Open(absPath), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.AutoMigrate(&Page{}, &Link{}, &WordCount{}); err != nil {
		return nil, fmt.Errorf("failed to migrate database schema: %w", err)
	}

	return db, nil
}