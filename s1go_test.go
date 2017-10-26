package s1go

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"testing"
)

var (
	username = flag.String("username", "", "username")
	password = flag.String("password", "", "password")
	real     = flag.Bool("real", false, "Send real request to s1 backend.")
)

type MockTransport struct {
	resp *http.Response
	err  error
}

func (m *MockTransport) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	return m.resp, m.err
}

func mockS1Client(resp *http.Response, err error) *S1Client {
	if resp != nil && resp.Body == nil {
		resp.Body = ioutil.NopCloser(nil)
	}
	httpClient := &http.Client{
		Transport: &MockTransport{resp, err},
	}
	return &S1Client{
		httpClient: httpClient,
	}
}

func TestLoginError(t *testing.T) {
	s1client := mockS1Client(nil, errors.New(""))
	err := s1client.Login("abc", "bcd")
	if err == nil {
		t.Errorf("Should return error")
	}
}

func TestLoginFailed(t *testing.T) {
	s1client := mockS1Client(&http.Response{}, nil)
	err := s1client.Login("abc", "bcd")
	if err == nil {
		t.Errorf("Shold return error")
	}
}

func TestGetForum(t *testing.T) {
	if !*real {
		t.SkipNow()
	}

	s1client := NewS1Client()
	mayLogin(s1client)

	forums, _ := s1client.GetForums()
	if len(forums) == 0 {
		t.Error("Cannot find any forums")
	}
}

func TestGetThread(t *testing.T) {
	if !*real {
		t.SkipNow()
	}

	s1client := NewS1Client()
	forum := Forum{ID: 51}

	threads, err := s1client.GetThreads(forum, 0)
	if err != nil {
		t.Error(err)
	}

	if len(threads) == 0 {
		t.Error("Cannot find any threads")
	}
}

func TestParseThread(t *testing.T) {
	resp := mockHTTPResponse("<ul type=\"1\"><li><a href=\"tid-793032.html\">二手交易区开放，增加求购，发帖前必看版规，否则抹布，</a> (48篇回复)</li></ul>")
	s1client := mockS1Client(resp, nil)
	threads, err := s1client.GetThreads(Forum{ID: 51}, 0)

	if err != nil {
		t.Error(err)
	}

	if len(threads) == 0 {
		t.Error("Threads should not be empty")
	}

	if threads[0].ID != 793032 {
		t.Error()
	}

	if threads[0].Reply != 48 {
		t.Error()
	}
}

func mayLogin(client *S1Client) error {
	if len(*username) != 0 {
		return client.Login(*username, *password)
	}
	return nil
}

func mockHTTPResponse(content string) *http.Response {
	u, _ := url.Parse("http://www.google.com")
	return &http.Response{
		Body:    ioutil.NopCloser(strings.NewReader(content)),
		Request: &http.Request{URL: u},
	}
}

func TestRealLoginSuccess(t *testing.T) {
	if !*real {
		t.SkipNow()
	}

	if len(*username) == 0 || len(*password) == 0 {
		t.SkipNow()
	}

	s1client := NewS1Client()
	err := s1client.Login(*username, *password)
	if err != nil {
		t.Error(err)
	}

	for _, cookie := range s1client.cookies {
		fmt.Printf("Cookie %s: %s\n", cookie.Name, cookie.Value)
	}
}

func TestRealFetchPostSuccess(t *testing.T) {
	if !*real {
		t.SkipNow()
	}

	s1client := NewS1Client()
	threads, err := s1client.GetPosts(Thread{ID: 1559184}, 1)
	if err != nil {
		t.Error(err)
	}
	if len(threads) == 0 {
		t.Error("Emtpy Thread")
	}

	for _, thread := range threads {
		fmt.Printf("Thread: %s\n", *thread)
	}
}

func TestLoginSuccess(t *testing.T) {
	cookie := http.Cookie{Name: authCookie, Value: "123"}
	header := http.Header{}
	header.Add("Set-Cookie", cookie.String())
	resp := &http.Response{Header: header}
	s1client := mockS1Client(resp, nil)
	if err := s1client.Login("abc", "bcd"); err != nil {
		t.Error(err)
	}
}
