package storage

import (
	"fmt"
	"github.com/boltdb/bolt"
	"github.com/golang/protobuf/proto"
	"github.com/smy20011/s1go/stage1stpb"
)

var BUCKET = []byte("stage1st")

type Storage struct {
	db *bolt.DB
}

func (s *Storage) Get(threadId int) (thread *stage1stpb.Thread, err error) {
	thread = &stage1stpb.Thread{}
	err = s.db.View(func(tx *bolt.Tx) error {
		key := []byte(fmt.Sprint(threadId))
		bytes := tx.Bucket(BUCKET).Get(key)
		return proto.Unmarshal(bytes, thread)
	})
	return
}

func (s *Storage) Put(thread *stage1stpb.Thread) (err error) {
	return s.db.Update(func(tx *bolt.Tx) error {
		key := []byte(fmt.Sprint(thread.ThreadId))
		bytes, err := proto.Marshal(thread)
		if err != nil {
			return err
		}
		return tx.Bucket(BUCKET).Put(key, bytes)
	})
}

func (s *Storage) Close() {
	s.db.Close()
}

func Open(filename string) (Storage, error) {
	db, err := bolt.Open(filename, 0600, nil)
	if err != nil {
		return Storage{}, err
	}
	// Create default buckets.
	err = db.Update(func(tx *bolt.Tx) (err error) {
		_, err = tx.CreateBucketIfNotExists(BUCKET)
		return
	})
	if err != nil {
		db.Close()
		return Storage{}, err
	}
	return Storage{db}, nil
}
