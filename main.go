package main

import (
	"flag"
	"time"
	"github.com/smy20011/s1go/crawler"
	"os"
	"os/signal"
	"syscall"
	"log"
)

var (
	interval = flag.Int64("interval", 3600, "Seconds between different fetch")
	username = flag.String("username", "", "Stage1st username")
	password = flag.String("password", "", "Stage1st password")
)

func main() {
	flag.Parse()

	c, err := crawler.NewCrawler()
	if err != nil {
		panic(err)
	}
	c.StartQueryServer()
	defer c.Close()
	trapCtrlCAndClose(c)
	if len(*username) > 0 {
		c.Login(*username, *password)
	}

	trigger := time.Tick(time.Second * time.Duration(*interval))
	for {
		c.FetchAllForums()
		<-trigger
	}
}

func trapCtrlCAndClose(c *crawler.Crawler) {
	channel := make(chan os.Signal, 2)
	signal.Notify(channel, os.Interrupt, syscall.SIGTERM)
	go func() {
		<- channel
		log.Printf("Gracefully shutdown!")
		c.Close()
		os.Exit(0)
	}()
}
