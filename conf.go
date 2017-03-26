package fund_crawler

import (
	"fmt"
	"os"
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

type Fund struct {
	gorm.Model
	Symbol    string
	Name      string
	Type      string
	Available bool     `gorm:"default:true"`
	Done      bool     `gorm:"default:false"`
	Records   []Record `gorm:"ForeignKey:FundID"`
	CAGR      float32
}

type Record struct {
	gorm.Model
	Day    time.Time
	Open   int // Cents
	Close  int // Cents
	FundID uint
}

// Manually set the Fund's table name to the sample we created.
// func (Fund) TableName() string {
// 	return "sampled_funds"
// }

func GetDB() *gorm.DB {
	var adapter string
	var dbPath string
	if os.Getenv("CLOUD_BABY") == "YEAH_BABY" {
		fmt.Println("We're in the cloud, baby")
		adapter = "mysql"
		dbPath = "pink:Tbz7vr2yiiaywNHF6Uu@/index_funds3?charset=utf8&parseTime=True&loc=Local"
	} else {
		fmt.Println("We're running locally, baby")
		adapter = "sqlite3"
		dbPath = "db/funds.db"
	}
	db, err := gorm.Open(adapter, dbPath)

	if err != nil {
		panic("failed to connect database")
	}
	db.DB().SetMaxIdleConns(100)
	return db
}
