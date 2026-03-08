package main

import (
	"log"

	"crawler/crawl"
	"crawler/persistence"
	"crawler/util"
)

func main() {

	config, err := util.LoadConfig("../config.yml")
	if err != nil {
		log.Fatal(err)
	}

	db, err := persistence.InitalizeDatabase(
		config.SearchEngine.Database.StoragePath,
	)
	if err != nil {
		log.Fatal(err)
	}

	crawler := crawl.NewCrawler(
		db,
		config.SearchEngine.Crawler.StartURL,
		"fruitsA",
		config.SearchEngine.Crawler.CrawlLimit,
		config.SearchEngine.Crawler.MaxWorkers,
	)

	if err := crawler.Crawl(); err != nil {
		log.Fatal(err)
	}

	log.Println("crawler finished successfully")
}