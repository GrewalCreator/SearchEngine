package main

import (
	"fmt"
	"log"
	"sort"

	"crawler/crawl"
	"crawler/parser"
)

func main() {
	pageURL := "https://people.scs.carleton.ca/~avamckenney/fruitsA/N-0.html"

	body, err := crawl.FetchPage(pageURL)
	if err != nil {
		log.Fatal(err)
	}

	parsed, err := parser.ParsePage(pageURL, body)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("==== PARSER TEST ====")
	fmt.Println("URL:", parsed.URL)
	fmt.Println("Title:", parsed.Title)
	fmt.Println("Total Links:", len(parsed.Links))
	fmt.Println("Total Unique Words:", len(parsed.WordCounts))
	fmt.Println()

	fmt.Println("First 10 Links:")
	linkLimit := min(10, len(parsed.Links))
	for i := 0; i < linkLimit; i++ {
		fmt.Printf("%d. %s\n", i+1, parsed.Links[i])
	}
	fmt.Println()

	fmt.Println("Sample Word Counts:")
	words := make([]string, 0, len(parsed.WordCounts))
	for word := range parsed.WordCounts {
		words = append(words, word)
	}
	sort.Strings(words)

	wordLimit := min(10, len(words))
	for i := 0; i < wordLimit; i++ {
		word := words[i]
		fmt.Printf("%s: %d\n", word, parsed.WordCounts[word])
	}
}

