package main

import (
	"fund_crawler"
	"log"

	"net/http"
	_ "net/http/pprof"
)

// catch interrupt signals: http://stackoverflow.com/a/18158859
func main() {
	// c := make(chan os.Signal, 2)
	// signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	// go func() {
	// 	<-c
	// 	fmt.Println("Cleaning up")
	// 	os.Exit(1)
	// }()
	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()
	fund_crawler.Crawl()
}
