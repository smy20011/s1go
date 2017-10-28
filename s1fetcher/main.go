package main

import (
	"flag"
	"github.com/smy20011/s1go"
	"log"
	"strconv"
	"strings"
	"time"
)

var (
	forumIds = flag.String("forum_ids", "", "Comma-separated forum ids, empty means all forum")
	interval = flag.Int64("interval", 3600, "Seconds between different fetch")
)

func main() {
	flag.Parse()

	crawler := NewCrawler()
	defer crawler.Close()

	forums := getForums(crawler)
	trigger := time.Tick(time.Second * time.Duration(*interval))
	for {
		for _, forum := range forums {
			log.Printf("Start Fetch Forum %s(%d)", forum.Title, forum.ID)
			go crawler.FetchForum(forum)
		}
		<-trigger
	}
}

func getForums(crawler Crawler) (result []s1go.Forum) {
	forums, err := crawler.S1Client.GetForums()
	panicIfErr(err)

	if len(*forumIds) == 0 {
		return forums
	}

	forumMap := map[int]s1go.Forum{}
	for _, forum := range forums {
		forumMap[forum.ID] = forum
	}
	for _, idStr := range strings.Split(*forumIds, ",") {
		id, err := strconv.Atoi(idStr)
		panicIfErr(err)
		result = append(result, forumMap[id])
	}
	return result
}
