package main

import (
	"database/sql"
	"flag"
	_ "github.com/mattn/go-sqlite3"
	"github.com/smy20011/s1go/client"
	"log"
	"sync"
	"time"
)

const (
	postSchema = `
	CREATE TABLE IF NOT EXISTS S1Posts (
		id              INTEGER PRIMARY KEY,
		fetch_time_msec INTEGER,
		thread_id       INTEGER,
		forum_id        INTEGER,
		post_index      INTEGER,
		post_time_msec  INTEGER,
		author          TEXT,
		content         TEXT
	)
	`
	threadSchema = `
	CREATE TABLE IF NOT EXISTS S1Threads (
		id              INTEGER PRIMARY KEY,
		thread_id       INTEGER,
		forum_id        INTEGER,
		fetch_time_msec INTEGER,
		thread_index    INTEGER,
		title           TEXT,
		reply           INTEGER
	)
	`
	postsPerPage = 30
)

var (
	username      = flag.String("username", "", "Stage1st Username")
	password      = flag.String("password", "", "Stage1st Password")
	pagePerForum  = flag.Int("page_per_forum", 2, "Craw # pages per forum")
	pagePerThread = flag.Int("page_per_thread", 2, "Craw # pages per thread")
	dbFile        = flag.String("db", "stage1st.db", "Location to store stage1st database")
)

// NewCrawler creates a new Crawler object using command line flags.
// Make sure call Crawler.Close() after finish.
func NewCrawler() Crawler {
	// Open database
	db, err := sql.Open("sqlite3", *dbFile)
	panicIfErr(err)

	// Create s1 client and login if possible.
	client := client.NewS1Client()
	if len(*username) != 0 && len(*password) != 0 {
		err := client.Login(*username, *password)
		panicIfErr(err)
	}

	crawler := Crawler{
		db, client, *pagePerForum, *pagePerThread, &sync.Mutex{},
	}

	// Create databases
	crawler.executeSql(postSchema)
	crawler.executeSql(threadSchema)
	return crawler
}

type Crawler struct {
	DB            *sql.DB
	S1Client      *client.S1Client
	PagePerForum  int
	PagePerThread int
	dbLoc         *sync.Mutex
}

func (c *Crawler) FetchForum(forum client.Forum) error {
	count := 0
	for page := 1; page <= c.PagePerForum; page++ {
		threads, err := c.S1Client.GetThreads(forum, page)
		if err != nil {
			log.Printf(
				"Error when fetching forum %v page %v: %v",
				forum, page, err)
			return err
		}

		for _, thread := range threads {
			count++
			c.saveThread(count, thread)

			// Avoid duplicate fetch of a thread.
			if c.isThreadFetched(thread) {
				continue
			}

			log.Printf(
				"Fetch new thread: %s\n",
				thread.Title)

			err = c.fetchThread(thread)
			if err != nil {
				return err
			}
		}
	}
	log.Printf("Fetch Forum %s Finished", forum.Title)
	return nil
}

func (c *Crawler) fetchThread(thread client.Thread) error {
	count := 0
	for page := 1; page <= c.PagePerThread; page++ {
		// Exit if reach end of thread.
		if postsPerPage*(page-1) > thread.Reply {
			return nil
		}

		posts, err := c.S1Client.GetPosts(thread, page)
		if err != nil {
			log.Printf(
				"Error when fetching thread %v page %v: %v",
				thread, page, err)
			return err
		}

		for _, post := range posts {
			count++
			c.savePost(count, post)
		}
	}
	return nil
}

func (c *Crawler) savePost(index int, post *client.Post) {
	c.executeSql(`
		INSERT INTO
			S1Posts(
				thread_id, forum_id, post_index,
				post_time_msec, author, content, fetch_time_msec)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		post.Thread.ID, post.Thread.Forum.ID, index,
		post.PostTime.Unix(), post.Author, post.Content,
		time.Now().Unix())
}

func (c *Crawler) saveThread(index int, thread client.Thread) {
	c.executeSql(`
		INSERT INTO S1Threads(
			thread_id, forum_id, thread_index,
			title, reply, fetch_time_msec)
		VALUES (?, ?, ?, ?, ?, ?)`,
		thread.ID, thread.Forum.ID, index,
		thread.Title, thread.Reply, time.Now().Unix())
}

func (c *Crawler) isThreadFetched(thread client.Thread) bool {
	c.dbLoc.Lock()
	result, err := c.DB.Query(`
		SELECT
			COUNT(*)
		FROM 
			S1Posts
		WHERE
			thread_id = ?`,
		thread.ID)
	c.dbLoc.Unlock()

	panicIfErr(err)
	defer result.Close()

	if result.Next() {
		var count int
		result.Scan(&count)
		return count > 0

	} else {
		return false
	}
}

func (c *Crawler) executeSql(query string, args ...interface{}) {
	c.dbLoc.Lock()
	_, err := c.DB.Exec(query, args...)
	c.dbLoc.Unlock()
	if err != nil {
		panic(err)
	}
}

func (c *Crawler) Close() error {
	return c.DB.Close()
}

func panicIfErr(err error) {
	if err != nil {
		panic(err)
	}
}
