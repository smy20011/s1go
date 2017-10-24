package s1go

import (
	"errors"
	"net/http"
	"net/url"
	"time"
)

const baseUrl = "https://bbs.saraba1st.com/2b/"
const loginUrl = "member.php?mod=logging&action=login&loginsubmit=yes&infloat=yes&lssubmit=yes&inajax=1"
const frontPage = "archiver/"
const forumUrlTemplate = "archiver/fid-%d.html?page=%d"
const threadUrlTemplate = "archiver/tid-%d.html?page=%d"
const authCookie = "B7Y9_2132_auth"

type Forum struct {
	Id        int
	PageCount int
}

type Thread struct {
	Id        int
	PageCount int
}

type Post struct {
	Author   string
	PostTime time.Time
	Content  string
}

type S1Client struct {
	httpClient *http.Client
	cookies    []*http.Cookie
}

func NewS1Client() *S1Client {
	return &S1Client{
		httpClient: &http.Client{},
	}
}

func (s *S1Client) login(username string, password string) (err error) {
	data := url.Values{
		"username":       {username},
		"password":       {password},
		"fastloginfield": {"username"},
		"quickforward":   {"yes"},
		"handlekey":      {"ls"},
	}
	resp, err := s.httpClient.PostForm(baseUrl+loginUrl, data)
	if err != nil {
		return err
	}

	for _, cookie := range resp.Cookies() {
		if cookie.Name == authCookie {
			s.cookies = resp.Cookies()
			return nil
		}
	}
	return errors.New("Login Failed")
}
