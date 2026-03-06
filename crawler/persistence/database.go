package persistence

import (
	"fmt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

/*
- InitalizeDatabase inits the database using gorm
- @param  dbPath: Path to data storage
- @return *gorm.DB: Pointer to Database object
- @return error occuring during initalization
*/
func InitalizeDatabase(dbPath string)(*gorm.DB, error){
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect database: %w", err)
  	}

	// Migrate the schema
  	err = db.AutoMigrate(&Page{}, &Link{}, &WordCount{})
	if err != nil {
		return nil, fmt.Errorf("failed to migrate schema: %w", err)
	}

	return db, nil

}