package fund_crawler

import (
	"fmt"
	"os"
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
	WG    *sync.WaitGroup
	Funds chan Fund
}

func ScrapeRecords(state *CrawlerState) {
	select {
	case fund := <-state.Funds:
		fund.PopulateRecords(state.DB)
	default:
		state.WG.Done()
	}
}

// Main function that controls the crawler.
func Crawl() {
	var adapter string
	var dbPath string
	if os.Getenv("CLOUD_BABY") == "YEAH_BABY" {
		fmt.Println("We're in the cloud, baby")
		adapter = "mysql"
		dbPath = "pink:Tbz7vr2yiiaywNHF6Uu@/index_funds2?charset=utf8&parseTime=True&loc=Local"
	} else {
		fmt.Println("We're running locally, baby")
		adapter = "sqlite3"
		dbPath = "db/funds.db"
	}
	db, err := gorm.Open(adapter, dbPath)
	if err != nil {
		panic("failed to connect database")
	}
	defer db.Close()

	// Migrate the schema
	db.AutoMigrate(&Fund{}, &Record{}, &AnnualReturn{})

	// Get funds to scrape historical data for
	allFunds := []Fund{}
	db.Where("done = 0").Find(&allFunds)

	// Set up worker goroutines and Funds channel
	state := CrawlerState{
		DB:    db,
		WG:    &sync.WaitGroup{},
		Funds: make(chan Fund, len(allFunds)),
	}

	fmt.Println("Funds to scrape: ", len(allFunds))
	for _, fund := range allFunds {
		state.Funds <- fund
	}

	// Fan out
	for i := 0; i < NumWorkers; i++ {
		state.WG.Add(1)
		go ScrapeRecords(&state)
	}

	state.WG.Wait()

	allFunds = []Fund{}
	db.Where("done_perf = 0").Find(&allFunds)
	for _, fund := range allFunds {
		records := []Record{}
		db.Where("fund_id = ?", fund.ID).Group("year(day), month(day)").Having("month(day) = 1 or month(day) = 12").Find(&records)
		fmt.Printf("%s (%d records)\n", fund.Symbol, len(records))

		if len(records)%2 != 0 || len(records) == 0 {
			fmt.Printf("Bad # rows (%d)\n", len(records))
			fund.BadData = true
			db.Save(&fund)
			continue
		}

		for i := 0; i < len(records)-1; i += 2 {
			year := records[i].Day.Year()
			yearOpening := records[i].Open
			yearClosing := records[i+1].Close
			var diff float64 = float64(yearClosing-yearOpening) / float64(yearOpening)
			fmt.Printf("\t%d: %.3f\n", year, diff)
			performance := AnnualReturn{FundID: fund.ID, Year: year, Diff: diff}
			db.Create(&performance)
		}
		fund.DonePerf = true
		db.Save(&fund)
	}
}
