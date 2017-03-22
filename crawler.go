package fund_crawler

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

const (
	CSVDateIndex  = 0
	CSVOpenIndex  = 1
	CSVCloseIndex = 4
)

type Fund struct {
	gorm.Model
	Symbol    string
	Name      string
	Type      string
	Available bool
	Records   []Record `gorm:"ForeignKey:FundID"`
}

type Record struct {
	gorm.Model
	Day    time.Time
	Open   float64
	Close  float64
	FundID uint
}

func (self *Fund) PopulateRecords(db *gorm.DB) {
	url := BuildQueryString(self)
	response := FetchCSV(url, self)
	records := ParseRecords(response, self)
	for _, record := range *records {
		db.Create(&record)
	}
}

func BuildQueryString(fund *Fund) *url.URL {
	u, err := url.Parse("http://ichart.finance.yahoo.com/table.csv?s=VOO&a=11&b=15&c=2000&d=11&e=19&f=2017&g=d&ignore=.csv")
	if err != nil {
		log.Fatal(err)
	}
	q := u.Query()
	q.Set("s", fund.Symbol)
	u.RawQuery = q.Encode()
	return u
}

func FetchCSV(url *url.URL, fund *Fund) *http.Response {
	response, err := http.Get(url.String())
	// if response != nil {
	// 	defer response.Body.Close()
	// }
	if err != nil {
		log.Fatal(err)
	}
	return response

	// records, err = reader.ReadAll()
	// if err != nil {
	// 	log.Fatal(err)
	// }
	return response
}

func ParseRecords(response *http.Response, fund *Fund) *[]Record {
	// Parse as CSV
	defer response.Body.Close()
	reader := csv.NewReader(bufio.NewReader(response.Body))

	var records []Record
	isFirstRecord := true
	for {
		record, err := reader.Read()

		if err == io.EOF {
			break
		}

		// skip parsing the table header
		if isFirstRecord {
			isFirstRecord = false
			continue
		}

		if err != nil {
			log.Fatal(err)
		}
		// fmt.Println(record)
		// for value := range record {
		// 	fmt.Printf("\t%v\n", record[value])
		// }

		// Convert prices from strings to floats
		openPrice, err := strconv.ParseFloat(record[CSVOpenIndex], 32)
		if err != nil {
			log.Fatal(err)
		}

		closePrice, err := strconv.ParseFloat(record[CSVCloseIndex], 32)
		if err != nil {
			log.Fatal(err)
		}

		// Convert time from string to time
		const dateFormat = "2006-01-02"
		recordDate, err := time.Parse(dateFormat, record[CSVDateIndex])
		if err != nil {
			log.Fatal(err)
		}

		var fundRecord = Record{
			Day:    recordDate,
			Open:   openPrice,
			Close:  closePrice,
			FundID: fund.ID}
		records = append(records, fundRecord)
	}
	return &records
}

func Crawl() {
	db, err := gorm.Open("sqlite3", "db/funds.db")
	if err != nil {
		panic("failed to connect database")
	}
	defer db.Close()

	// Migrate the schema
	db.AutoMigrate(&Fund{}, &Record{})

	// Create a single example Fund
	exampleFund := Fund{Symbol: "VOO", Name: "Vanguard S&P 500", Type: "etf", Available: true}
	if db.NewRecord(exampleFund) {
		db.Create(&exampleFund)
	}

	var fund Fund
	db.First(&fund)
	fund.PopulateRecords(db)

	//var records *[]Record = GetFundRecords(fund)

	// Create
	//db.Create(&Product{Code: "L1212", Price: 1000})

	// Read
	// var product Product
	// db.First(&product, 1)                   // find product with id 1
	// db.First(&product, "code = ?", "L1212") // find product with code l1212

	// Update - update product's price to 2000
	// db.Model(&product).Update("Price", 2000)

	// Delete - delete product
	// db.Delete(&product)
}
