package storage

import (
	"github.com/smy20011/s1go/stage1stpb"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"github.com/golang/protobuf/proto"
	"path"
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
	tmpFile := path.Join(tmpDir, "benchmark.DB")
	storage, _ := Open(tmpFile)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		thread := stage1stpb.Thread{ThreadId: int32(i)}
		storage.Put(&thread)
		storage.Get(i)
	}
	storage.Close()
	os.RemoveAll(tmpDir + "/")
}

func TestMain(m *testing.M) {
	tmpDir, _ = ioutil.TempDir("", "DB")
	retcode := m.Run()
	err := os.RemoveAll(tmpDir + "/")
	if err != nil {
		panic(err)
	}
	os.Exit(retcode)
}
