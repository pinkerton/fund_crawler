package fund_crawler

import (
	"bufio"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

func (self *Fund) CalculateReturn(db *gorm.DB) {
	records := []Record{}
	db.Where("fund_id = ?", self.ID).Group("year(day), month(day)").Having("month(day) = 1 or month(day) = 12").Find(&records)
	// fmt.Printf("%s (%d records)\n", self.Symbol, len(records))

	if len(records)%2 != 0 || len(records) == 0 {
		// fmt.Printf("Bad # rows (%d)\n", len(records))
		self.Available = false
		self.Done = true
		db.Save(&self)
		fmt.Printf("Skipping %s (bad return)\n", self.Symbol)
		return
	}

	for i := 0; i < len(records)-1; i += 2 {
		year := records[i].Day.Year()
		yearOpening := records[i].Open
		yearClosing := records[i+1].Close
		var diff float64 = float64(yearClosing-yearOpening) / float64(yearOpening)
		// fmt.Printf("\t%d: %.3f\n", year, diff)
		performance := AnnualReturn{FundID: self.ID, Year: year, Diff: diff}
		db.Create(&performance)
	}
}

// High-level method that calls functions to request, parse, and create Records.
func (self *Fund) PopulateRecords(db *gorm.DB) (err error) {
	url := BuildQueryString(self)
	response := FetchCSV(url, self)
	records, err := ParseRecords(response, self)
	if err != nil {
		fmt.Println(err)
		self.Available = false
		self.Done = true
		db.Save(&self)
		return
	}
	for _, record := range *records {
		db.Create(&record)
	}
	self.Done = true
	db.Save(&self)
	return
}

// Build the URL we'll GET with the specific fund's symbol.
// The time range is hardcoded: Jan. 1, 2000 to Dec. 31, 2016.
func BuildQueryString(fund *Fund) *url.URL {
	u, err := url.Parse("http://ichart.finance.yahoo.com/table.csv?s=VOO&a=00&b=01&c=2000&d=11&e=31&f=2016&g=d&ignore=.csv")
	if err != nil {
		log.Fatal(err)
	}
	q := u.Query()
	q.Set("s", fund.Symbol)
	u.RawQuery = q.Encode()
	return u
}

// Make a GET request to the built URL and return it.
// Assumes the caller will close response.Body.
// TODO: Consider the implicaitons of this. It makes more sense to do build a
// CSV reader and ReadAll() into a 2d array and return a pointer to that,
// because the caller should not have to clean up after us.
func FetchCSV(url *url.URL, fund *Fund) *http.Response {
	response, err := http.Get(url.String())
	if err != nil {
		log.Fatal(err)
	}
	return response
}

// Parse the response data as CSV, and create a new Record for each row.
// TODO: Refactor into smaller units.
// TODO: Add multithreading.
func ParseRecords(response *http.Response, fund *Fund) (*[]Record, error) {
	// Parse as CSV
	defer response.Body.Close()
	reader := csv.NewReader(bufio.NewReader(response.Body))
	records := make([]Record, 2000)
	var err error
	isFirstRecord := true

	for {
		record, err := reader.Read()

		if err == io.EOF {
			break
		}

		if err != nil {
			err = nil
			break
		}

		// skip parsing the csv table header
		if isFirstRecord {
			isFirstRecord = false
			continue
		}

		// Convert prices from strings to floats
		openPrice, err := strconv.ParseFloat(record[CSVOpenIndex], 32)
		if err != nil {
			err = errors.New("failed to parse open price")
		}
		openPriceCents := int(openPrice * 100)

		closePrice, err := strconv.ParseFloat(record[CSVCloseIndex], 32)
		if err != nil {
			err = errors.New("failed to parse close price")
		}
		closePriceCents := int(closePrice * 100)

		// Convert time from string to time
		const dateFormat = "2006-01-02"
		recordDate, err := time.Parse(dateFormat, record[CSVDateIndex])
		if err != nil {
			err = errors.New("failed to parse quote date")
		}

		var fundRecord = Record{
			Day:    recordDate,
			Open:   openPriceCents,
			Close:  closePriceCents,
			FundID: fund.ID}
		records = append(records, fundRecord)
	}
	return &records, err
}
