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
	"os"
	"strconv"
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
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
	Available bool     `gorm:"default:true"`
	Done      bool     `gorm:"default:false"`
	Records   []Record `gorm:"ForeignKey:FundID"`
}

type Record struct {
	gorm.Model
	Day    time.Time
	Open   int
	Close  int
	FundID uint
}

func (self *Fund) PopulateRecords(db *gorm.DB) {
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
	if err != nil {
		log.Fatal(err)
	}
	return response
}

func ParseRecords(response *http.Response, fund *Fund) (*[]Record, error) {
	// Parse as CSV
	defer response.Body.Close()
	reader := csv.NewReader(bufio.NewReader(response.Body))
	var records []Record
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

		// skip parsing the table header
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

func Crawl() {
	var adapter string
	var dbPath string
	if os.Getenv("CLOUD_BABY") == "YEAH_BABY" {
		adapter = "mysql"
		dbPath = "pink:Tbz7vr2yiiaywNHF6Uu@/index_funds?charset=utf8&parseTime=True&loc=Local"
	} else {
		adapter = "sqlite3"
		dbPath = "db/funds.db"
	}
	db, err := gorm.Open(adapter, dbPath)
	if err != nil {
		panic("failed to connect database")
	}
	defer db.Close()

	// Migrate the schema
	db.AutoMigrate(&Fund{}, &Record{})

	funds := []Fund{}
	db.Where(&Fund{Done: false}).Find(&funds)
	fmt.Println(len(funds))
	for _, fund := range funds {
		fmt.Println("Fetching fund: ", fund.Symbol)
		fund.PopulateRecords(db)
	}
}
