package crawler

import (
	"github.com/smy20011/s1go/client"
	"github.com/smy20011/s1go/stage1stpb"
	"github.com/smy20011/s1go/storage"
	"log"
	"sync"
	"time"
)

var (
	depth           = 3
	postPerPage     = 30
	maxThreadPage   = 3
	maxThreadUpdate = 500
)

type Crawler struct {
	S1Client *client.S1Client
	Storage  *storage.Storage
}

func NewCrawler(client *client.S1Client, storage *storage.Storage) *Crawler {
	return &Crawler{
		S1Client: client,
		Storage:  storage,
	}
}

func (c *Crawler) Login(username, password string) error {
	return c.S1Client.Login(username, password)
}

func (c *Crawler) FetchAllForums() error {
	forums, err := c.S1Client.GetForums()
	if err != nil {
		return err
	}
	wg := sync.WaitGroup{}
	for _, forum := range forums {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := c.fetchForum(forum)
			if err != nil {
				log.Printf("Error while fetch forum %s: %v", forum.Title, err)
			}
		}()
	}
	wg.Wait()
	return nil
}

func (c *Crawler) Close() {
	c.Storage.Close()
}

func (c *Crawler) fetchForum(forum client.Forum) (err error) {
	threads := []client.Thread{}
	for i := 0; i < depth; i++ {
		newThreads, err := c.S1Client.GetThreads(forum, i)
		if err != nil {
			return err
		}
		threads = append(threads, newThreads...)
	}
	for index, thread := range threads {
		err = c.fetchThread(index, thread)
		if err != nil {
			return
		}
	}
	return nil
}

func (c *Crawler) fetchThread(index int, thread client.Thread) error {
	savedThread, err := c.Storage.Get(thread.ID)
	if err != nil {
		return err
	}
	if savedThread.ThreadId != int32(thread.ID) {
		savedThread = &stage1stpb.Thread{
			ThreadId: int32(thread.ID),
			ForumId:  int32(thread.Forum.ID),
			Title:    thread.Title,
		}
	}
	// Skip thread update if we receive update for at least 100 times.
	if len(savedThread.ThreadInfos) >= maxThreadUpdate {
		return nil
	}

	savedThread.ThreadInfos = append(savedThread.ThreadInfos, &stage1stpb.ThreadInfo{
		Rank:      int32(index),
		Replies:   int32(thread.Reply),
		Timestamp: time.Now().Unix(),
	})

	posts, err := c.fetchNewPosts(thread, len(savedThread.Posts))
	// Ignore post fetch error since it could failed due to lack of permission.
	if err != nil {
		log.Printf("Fetch Thread failed %v", err)
	}
	for _, post := range posts {
		savedThread.Posts = append(savedThread.Posts, &stage1stpb.Post{
			Author:   post.Author,
			Content:  post.Content,
			PostTime: post.PostTime.Unix(),
		})
	}

	return c.Storage.Put(savedThread)
}

func (c *Crawler) fetchNewPosts(thread client.Thread, fetched int) (posts []*client.Post, err error) {
	for _, page := range getPagesToFetch(fetched, thread.Reply) {
		// +1 Because S1 use 1 as the first page of thread.
		p, err := c.S1Client.GetPosts(thread, page+1)
		if err != nil {
			return posts, err
		}
		// Append new posts.
		for index, post := range p {
			postIndex := page*postPerPage + index
			if postIndex >= fetched && postIndex < thread.Reply {
				posts = append(posts, post)
			}
		}
	}
	return
}

func getPagesToFetch(fetched, current int) (result []int) {
	for i := 0; i < current/postPerPage+1 && i < maxThreadPage; i++ {
		postStart := i * postPerPage
		if postStart >= fetched {
			result = append(result, i)
		}
	}
	return
}
