package fund_crawler

import (
	"fmt"
	"sync"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

const (
	CSVDateIndex  = 0
	CSVOpenIndex  = 1
	CSVCloseIndex = 4
	NumWorkers    = 5
)

// CrawlerState holds state shared by worker goroutines.
type CrawlerState struct {
	DB    *gorm.DB
	WG    *sync.WaitGroup
	Funds chan *Fund
}

func ScrapeRecords(state *CrawlerState) {
	for fund := range state.Funds {
		fund.PopulateRecords(state.DB)
		fund.CalculateReturn(state.DB)
		fmt.Printf("%s\n", fund.Symbol)
	}
	fmt.Println("Done!")
	state.WG.Done()
}

// Main function that controls the crawler.
func Crawl() {
	db := GetDB()
	defer db.Close()

	// Migrate the schema
	db.AutoMigrate(&Fund{}, &Record{}, &AnnualReturn{})

	// Get funds to scrape historical data for
	allFunds := []Fund{}
	db.Where("done = 0", "available = 1").Find(&allFunds)

	// Set up worker goroutines and Funds channel
	state := CrawlerState{
		DB:    db,
		WG:    &sync.WaitGroup{},
		Funds: make(chan *Fund, len(allFunds)),
	}

	fmt.Println("Funds to scrape: ", len(allFunds))
	for _, fund := range allFunds {
		state.Funds <- &fund
	}
	close(state.Funds)

	// Fan out
	for i := 0; i < NumWorkers; i++ {
		state.WG.Add(1)
		go ScrapeRecords(&state)
	}
	state.WG.Wait()
}
