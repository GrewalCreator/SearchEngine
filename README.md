# Search Engine

A small search engine built in Go for COMP4601.

It includes:

- a crawler that visits pages and extracts content
- a SQLite-backed storage layer
- PageRank computation over the crawled link graph
- a REST API for searching datasets and viewing page details
- a simple frontend served by the backend

## Features

- Crawl and store pages from configured datasets
- Extract page titles, links, and word counts
- Compute PageRank from the page link graph
- Search by dataset with optional PageRank boosting
- Return page details including incoming links, outgoing links, and word counts
- Serve a browser frontend from the Go server

## Project Structure

```text
.
├── backend
│   ├── crawl
│   │   ├── crawler.go
│   │   └── fetch.go
│   ├── d-graph
│   │   └── pagerank.go
│   ├── go.mod
│   ├── go.sum
│   ├── main.go
│   ├── parser
│   │   └── parse.go
│   ├── persistence
│   │   ├── api.go
│   │   ├── database.go
│   │   └── models.go
│   ├── search
│   │   └── search.go
│   ├── server
│   │   └── server.go
│   └── util
│       └── config.go
├── config.yml
├── data
│   └── search_data.db
├── frontend
│   ├── index.html
│   ├── script.js
│   └── style.css
└── README.md
```

## Setup

### Install Go

#### Arch
```text
sudo pacman -S go
```

#### Ubuntu / Debian
```text
sudo apt install golang-go
```

### Setup Project
```text
cd backend
go mod init crawler
go mod tidy
```

### Run Server
```text
cd backend
go run main.go
```