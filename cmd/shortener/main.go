package main

import (
	"log"

	_ "net/http/pprof"
)

// go build -ldflags "-X main.buildVersion=v1.0.1 -X 'main.buildDate=$(date +'%Y/%m/%d %H:%M:%S')' -X 'main.buildCommit=$(git rev-parse HEAD)'" ./cmd/shortener
var (
	buildVersion string
	buildDate    string
	buildCommit  string
)

func main() {
	logInfo()

	if err := run(); err != nil {
		log.Fatalf("application error: %v", err)
	}
}

func logInfo() {
	if buildVersion == "" {
		buildVersion = "N/A"
	}
	if buildDate == "" {
		buildDate = "N/A"
	}
	if buildCommit == "" {
		buildCommit = "N/A"
	}

	log.Println("Build version: ", buildVersion)
	log.Println("Build date: ", buildDate)
	log.Println("Build commit: ", buildCommit)
}
