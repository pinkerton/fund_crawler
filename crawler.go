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
	NumWorkers    = 10
)

// CrawlerState holds state shared by worker goroutines.
type CrawlerState struct {
	DB    *gorm.DB
	WG    sync.WaitGroup
	Funds chan Fund
}

func ScrapeRecords(id int, state *CrawlerState) {
	for fund := range state.Funds {
		fmt.Printf("#%d: %s\n", id, fund.Symbol)

		// Get first and last records for the fund
		before, after, err := fund.GetRecords(state.DB)
		if err != nil {
			fmt.Printf("#%d\t=> Skipping %s (%s)\n", id, fund.Symbol, err)
			fund.Ignore()
			state.DB.Save(&fund)
			continue
		}

		// Calculate CAGR
		fund.CalculateReturn(before, after)

		// Done. Save everything.
		fund.Done = true
		state.DB.Save(&fund) // TODO: will this actuall save the fund?
		state.DB.Create(before)
		state.DB.Create(after)
		fmt.Printf("#%d\t=> Done (%s)\n", id, fund.Symbol)

	}
	fmt.Printf("#%d => Done!\n", id)
	state.WG.Done()
}

// Main function that controls the crawler.
func Crawl() {
	db := GetDB()
	defer db.Close()

	// Migrate the schema
	db.AutoMigrate(&Fund{}, &Record{})

	// Get funds to scrape historical data for
	allFunds := []Fund{}
	db.Where("done = 0 AND available = 1").Order("id asc").Find(&allFunds)

	// Set up worker goroutines and Funds channel
	state := CrawlerState{
		DB:    db,
		WG:    sync.WaitGroup{},
		Funds: make(chan Fund, len(allFunds)),
	}

	fmt.Println("Funds to scrape: ", len(allFunds))
	for _, fund := range allFunds {
		state.Funds <- fund
	}
	close(state.Funds)

	// Fan out
	for i := 0; i < NumWorkers; i++ {
		state.WG.Add(1)
		go ScrapeRecords(i, &state)
	}
	state.WG.Wait()
}
