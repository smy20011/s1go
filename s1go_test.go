package s1go

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
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

func TestGetForum_RealRequestIssued(t *testing.T) {
	if !*real {
		t.SkipNow()
	}

	s1client := NewS1Client()
	if len(*username) != 0 {
		s1client.Login(*username, *password)
	}

	forums, _ := s1client.GetForums()
	if len(forums) == 0 {
		t.Error("Cannot find any forums")
	}

	fmt.Println(forums)
}

func TestRealLoginSuccess(t *testing.T) {
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
