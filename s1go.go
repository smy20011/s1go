package s1go

import (
	"errors"
	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/publicsuffix"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"regexp"
	"strconv"
	"time"
)

const baseURL = "https://bbs.saraba1st.com/2b/"
const loginURL = baseURL + "member.php?mod=logging&action=login&loginsubmit=yes&infloat=yes&lssubmit=yes&inajax=1"
const frontPageURL = baseURL + "archiver/"
const forumURLTemplate = baseURL + "archiver/fid-%d.html?page=%d"
const threadURLTemplate = baseURL + "archiver/tid-%d.html?page=%d"
const authCookie = "B7Y9_2132_auth"

// Forum represents a S1 Forum.
type Forum struct {
	Title string
	ID    int
}

// Thread represents a discussion thread in a forum.
type Thread struct {
	Title string
	ID    int
	Forum Forum
}

// Post represents a post in a thread.
type Post struct {
	Author   string
	PostTime time.Time
	Content  string
}

// S1Client helps us interact with s1 backend with a presistant cookie.
type S1Client struct {
	httpClient *http.Client
	cookies    []*http.Cookie
}

// NewS1Client creates a new S1 client.
func NewS1Client() *S1Client {
	jar, err := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	if err != nil {
		panic(err)
	}

	httpClient := &http.Client{Jar: jar}
	return &S1Client{httpClient: httpClient}
}

// Login retrieve S1 Cookie by simulate user's login. S1Client will use the same
// cookie for following requests.
func (s *S1Client) Login(username string, password string) (err error) {
	data := url.Values{
		"username":       {username},
		"password":       {password},
		"fastloginfield": {"username"},
		"quickforward":   {"yes"},
		"handlekey":      {"ls"},
	}

	resp, err := s.httpClient.PostForm(loginURL, data)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	for _, cookie := range resp.Cookies() {
		if cookie.Name == authCookie {
			s.cookies = resp.Cookies()
			return nil
		}
	}
	return errors.New("Login Failed")
}

// GetForums returns forums that are visiable to this user.
func (s *S1Client) GetForums() (forums []Forum, err error) {
	doc, err := s.getAndParase(frontPageURL)
	if err != nil {
		return
	}

	forumNodes := doc.Find("#content a")
	for i := range forumNodes.Nodes {
		node := forumNodes.Eq(i)

		link, found := node.Attr("href")
		if !found {
			err = errors.New("Cannot find forum link")
			return
		}

		forum := Forum{
			Title: node.Text(),
			ID:    findIntAndParse(link),
		}
		forums = append(forums, forum)
	}
	return
}

func (s *S1Client) getAndParase(url string) (doc *goquery.Document, err error) {
	resp, err := s.httpClient.Get(url)
	if err != nil {
		return
	}
	return goquery.NewDocumentFromResponse(resp)
}

func findIntAndParse(s string) (result int) {
	pattern, _ := regexp.Compile("\\d+")
	intStr := pattern.FindString(s)
	if len(intStr) == 0 {
		panic("Cannot find integer in " + s)
	}
	result, _ = strconv.Atoi(intStr)
	return
}
