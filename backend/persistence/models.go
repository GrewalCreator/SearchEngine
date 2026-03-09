package persistence

type Page struct {
	ID       uint    `gorm:"primaryKey"`
	URL      string  `gorm:"not null;index:idx_url_dataset,unique"`
	Dataset  string  `gorm:"not null;index:idx_url_dataset,unique;index"`
	Title    string
	PageRank float64 `gorm:"default:0"`

	OutgoingLinks []Link      `gorm:"foreignKey:FromPageID"`
	IncomingLinks []Link      `gorm:"foreignKey:ToPageID"`
	WordCounts    []WordCount `gorm:"foreignKey:PageID"`
}

type Link struct {
	ID         uint 	`gorm:"primaryKey"`
	FromPageID uint 	`gorm:"index;not null"`
	ToPageID   uint 	`gorm:"index;not null"`
}

type WordCount struct {
	ID     uint   		`gorm:"primaryKey"`
	PageID uint			`gorm:"index;not null"`
	Word   string 		`gorm:"index;not null"`
	Count  uint			`gorm:"not null"`
}
