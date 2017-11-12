package crawler

import (
	"github.com/smy20011/s1go/client"
	"github.com/smy20011/s1go/storage"
	"github.com/smy20011/s1go/test_util"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"testing"
	"fmt"
)

func CreateMockS1Client() *client.S1Client {
	return &client.S1Client{
		HttpClient: &http.Client{
			Transport: &test_util.MockS1Website{},
		},
	}
}

func GetTempDir() string {
	p, err := ioutil.TempDir("", "DB")
	if err != nil {
		panic(err)
	}
	return p
}

type TestFixture struct {
	dir, file string
	crawler   *Crawler
}

func CreateTestFixture() *TestFixture {
	f := TestFixture{}
	f.dir = GetTempDir()
	f.file = path.Join(f.dir, "test.db")
	s, err := storage.Open(f.file)
	if err != nil {
		panic(err)
	}
	f.crawler = &Crawler{CreateMockS1Client(), &s}
	return &f
}

func (f *TestFixture) Cleanup() {
	f.crawler.Close()
	os.RemoveAll(f.dir + "/")
}

func TestCrawler_fetchThread(t *testing.T) {
	f := CreateTestFixture()
	defer f.Cleanup()
	f.crawler.fetchThread(12345, client.Thread{ID: 12345, Reply: 20})
	thread, err := f.crawler.Storage.Get(12345)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, int32(12345), thread.ThreadId)
	assert.Equal(t, 20, len(thread.Posts))
	assert.Equal(t, 1, len(thread.ThreadInfos))
}

func TestCrawler_fetchThread_maxPost(t *testing.T) {
	f := CreateTestFixture()
	defer f.Cleanup()
	f.crawler.fetchThread(12345, client.Thread{ID: 12345, Reply: 2000})
	f.crawler.fetchThread(12345, client.Thread{ID: 12345, Reply: 2000})
	f.crawler.fetchThread(12345, client.Thread{ID: 12345, Reply: 2000})
	thread, _ := f.crawler.Storage.Get(12345)
	assert.Equal(t, int32(12345), thread.ThreadId)
	assert.Equal(t, 90, len(thread.Posts))
	assert.Equal(t, 3, len(thread.ThreadInfos))
}

func TestCrawler_fetchThread_newPosts(t *testing.T) {
	f := CreateTestFixture()
	defer f.Cleanup()
	f.crawler.fetchThread(12345, client.Thread{ID:12345, Reply: 30})
	thread, _ := f.crawler.Storage.Get(12345)
	assert.Equal(t, 30, len(thread.Posts))

	f.crawler.fetchThread(12345, client.Thread{ID:12345, Reply: 40})
	thread, _ = f.crawler.Storage.Get(12345)
	assert.Equal(t, 40, len(thread.Posts))
	assert.Equal(t, 2, len(thread.ThreadInfos))
}

func TestCrawler_FetchAllForums(t *testing.T) {
	f := CreateTestFixture()
	defer f.Cleanup()
	maxThreadPage = 1
	depth = 1
	f.crawler.FetchAllForums()
}

func TestCrawler_StorageSize(t *testing.T) {
	t.SkipNow()
	f := CreateTestFixture()
	defer f.Cleanup()

	for i := 1 ; i < 500 ; i++ {
		f.crawler.fetchThread(i, client.Thread{ID: i, Reply: 100})
	}
	info, _ := os.Stat(f.file)
	f.crawler.Storage.Close()
	fmt.Printf("Storage Size for 500 threads: %v MB\n", float32(info.Size()) / 1024.0 / 1024.0)
}
