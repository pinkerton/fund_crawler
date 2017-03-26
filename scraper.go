package fund_crawler

import (
	"bufio"
	"encoding/csv"
	"errors"
	"log"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

const (
	FloatSize    = 32
	HoursPerYear = 24 * 365
)

// Calculate the compound annual growth rate for a fund between two Records.
// http://www.investinganswers.com/financial-dictionary/investing/compound-annual-growth-rate-cagr-1096
func (self *Fund) CalculateReturn(before *Record, after *Record) {
	hoursBtwn := after.Day.Sub(before.Day).Hours() // Duration type can represent up to 290 years. We're okay.
	yearsBtwn := float64(hoursBtwn) / HoursPerYear
	quotient := float64(after.Close) / float64(before.Open)
	cagr := math.Pow(quotient, 1/yearsBtwn) - 1
	self.CAGR = float32(cagr)
}

// High-level method that calls functions to request, parse, and create Records.
func (self *Fund) GetRecords(db *gorm.DB) (*Record, *Record, error) {
	url := BuildQueryString(self)
	response := FetchCSV(url, self)
	before, after, err := self.ParseRecords(response)
	if err != nil {
		// fmt.Println(err)
		// err = errors.New("unable to parse records")
		return nil, nil, err
	}
	return before, after, err
}

// Mark a fund as having bad data.
// Assumes the caller will call db.Save(self)
func (self *Fund) Ignore() {
	self.Available = false
	self.Done = true
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
func FetchCSV(url *url.URL, fund *Fund) *http.Response {
	response, err := http.Get(url.String())
	if err != nil {
		log.Fatal(err)
	}
	return response
}

// Convert an array of CSV fields into a Record type.
// Does not add the Fund ID or Create the value in the database.
func CSVToRecord(fields []string) (record Record, err error) {
	// Convert prices from strings to floats
	openPrice, err := strconv.ParseFloat(fields[CSVOpenIndex], FloatSize)
	if err != nil {
		err = errors.New("failed to parse open price")
	}
	openPriceCents := int(openPrice * 100)

	closePrice, err := strconv.ParseFloat(fields[CSVCloseIndex], FloatSize)
	if err != nil {
		err = errors.New("failed to parse close price")
	}
	closePriceCents := int(closePrice * 100)

	// Convert time from string to time
	const dateFormat = "2006-01-02"
	recordDate, err := time.Parse(dateFormat, fields[CSVDateIndex])
	if err != nil {
		err = errors.New("failed to parse quote date")
	}
	record.Day = recordDate
	record.Open = openPriceCents
	record.Close = closePriceCents
	return
}

// Parse the response data as CSV, and create a new Record for each row.
// TODO: Refactor into smaller units.
func (self *Fund) ParseRecords(response *http.Response) (*Record, *Record, error) {
	// Parse as CSV
	defer response.Body.Close()
	reader := csv.NewReader(bufio.NewReader(response.Body))

	csvRecords, err := reader.ReadAll()
	if err != nil {
		return nil, nil, err
	}

	lastRecordFields := csvRecords[1]
	firstRecordFields := csvRecords[len(csvRecords)-1]

	before, err := CSVToRecord(firstRecordFields)
	if err != nil {
		return nil, nil, err
	}

	after, err := CSVToRecord(lastRecordFields)
	if err != nil {
		return nil, nil, err
	}

	before.FundID = self.ID
	after.FundID = self.ID
	return &before, &after, err
}
