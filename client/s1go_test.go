package client

import (
	"flag"
	"github.com/smy20011/s1go/test_util"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

var (
	username = flag.String("user", "", "username")
	password = flag.String("pass", "", "password")
)

func CreateMockS1Client() *S1Client {
	return &S1Client{
		HttpClient: &http.Client{
			Transport: &test_util.MockS1Website{},
		},
	}
}

func TestGetForums(t *testing.T) {
	client := CreateMockS1Client()
	forums, err := client.GetForums()
	if err != nil {
		t.Error(err)
	}
	if len(forums) != 16 {
		t.Errorf("Wrong Forums %v", forums)
	}
}

func TestGetTheads(t *testing.T) {
	client := CreateMockS1Client()
	threads, err := client.GetThreads(Forum{ID: 1}, 0)
	if err != nil {
		t.Error(err)
	}
	if len(threads) != 51 {
		t.Errorf("Wrong Threads Count %v", len(threads))
	}
}

func TestGetPosts(t *testing.T) {
	client := CreateMockS1Client()
	posts, err := client.GetPosts(Thread{ID: 1}, 0)
	if err != nil {
		t.Error(err)
	}
	if len(posts) != 30 {
		t.Errorf("Wrong Posts Count %v", len(posts))
	}
}

func TestGetSinglePost(t *testing.T) {
	client := CreateMockS1Client()
	posts, err := client.GetPosts(Thread{
		ID: test_util.SinglePostThread,
	}, 0)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(posts))
}

func TestLogin(t *testing.T) {
	if len(*username) == 0 {
		t.SkipNow()
	}
	client := NewS1Client()
	// Test Failed login
	err := client.Login(*username, "WrongPassword")
	if err == nil {
		t.Fatalf("Expect login failed but no error returned\n")
	}
	// Test Success login
	err = client.Login(*username, *password)
	if err != nil {
		t.Fatalf("Expect login success but login failed: %v\n", err)
	}
}
