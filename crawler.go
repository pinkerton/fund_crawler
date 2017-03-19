package fund_crawler

import (
	"fmt"
	"log"
	"net/url"
)

func GetFundInfo(symbol string) {
	u, err := url.Parse("http://ichart.finance.yahoo.com/table.csv?s=VOO&a=11&b=15&c=2000&d=11&e=19&f=2017&g=d&ignore=.csv")
	if err != nil {
		log.Fatal(err)
	}
	q := u.Query()
	q.Set("s", symbol)
	u.RawQuery = q.Encode()
	fmt.Println(u)
}

func Run() {
	fmt.Println("asdf")

	// 1. Parse CSV (DONE)
	//    Iterate over all rows
	// 2. Make request (DONE)
	// 3. Parse data
	// 4. Store data
}
