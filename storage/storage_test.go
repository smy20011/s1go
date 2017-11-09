package storage

import (
	"github.com/golang/protobuf/proto"
	"github.com/smy20011/s1go/stage1stpb"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"sync/atomic"
	"testing"
)

var (
	tmpDir = ""
	thread = stage1stpb.Thread{ThreadId: 12345}
)

func TestStorage(t *testing.T) {
	storage, err := Open(filepath.Join(tmpDir, "get.DB"))
	if err != nil {
		t.Fatal(err)
	}
	defer storage.Close()
	if err := storage.Put(&thread); err != nil {
		t.Fatalf("Storage Put Failed: %v", err)
	}
	stored_thread, err := storage.Get(12345)
	if err != nil {
		t.Fatal(err)
	}
	if !proto.Equal(&thread, stored_thread) {
		t.Fatalf("Data not same! Expected %v, Actuall %v", thread, stored_thread)
	}
}

func BenchmarkStorage(b *testing.B) {
	tmpDir, _ = ioutil.TempDir("", "DB")
	defer os.RemoveAll(tmpDir + "/")
	tmpFile := path.Join(tmpDir, "benchmark.DB")
	storage, _ := Open(tmpFile)
	var counter int32 = 0
	b.SetParallelism(100)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			threadId := atomic.AddInt32(&counter, 1)
			thread := stage1stpb.Thread{ThreadId: threadId}
			storage.Put(&thread)
			storage.Get(int(threadId))
		}
	})
	storage.Close()
}

func ExecTest(m *testing.M) (retcode int) {
	tmpDir, _ = ioutil.TempDir("", "DB")
	defer os.RemoveAll(tmpDir + "/")
	return m.Run()
}

func TestMain(m *testing.M) {
	os.Exit(ExecTest(m))
}
