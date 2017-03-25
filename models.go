package fund_crawler

import (
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
	BadData   bool     `gorm:"default:false"`
	DonePerf  bool     `gorm:"default:false"`
	Records   []Record `gorm:"ForeignKey:FundID"`
}

type Record struct {
	gorm.Model
	Day    time.Time
	Open   int
	Close  int
	FundID uint
}

type AnnualReturn struct {
	gorm.Model
	FundID uint
	Year   int
	Diff   float64
}

// Manually set the Fund's table name to the sample we created.
func (Fund) TableName() string {
	return "sampled_funds"
}
