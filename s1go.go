package s1go

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html"
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
	Reply int
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

// GetThreads returns threads in some forum at some page.
func (s *S1Client) GetThreads(forum Forum, page int) (threads []Thread, err error) {
	url := fmt.Sprintf(forumURLTemplate, forum.ID, page)

	doc, err := s.getAndParase(url)
	if err != nil {
		return
	}

	nodes := doc.Find("ul[type] li")
	for i := range nodes.Nodes {
		node := nodes.Eq(i)
		linkNode := node.Find("a")

		link, exists := linkNode.Attr("href")
		if !exists {
			err = errors.New("Cannot find thread link")
			return
		}

		thread := Thread{
			Forum: forum,
			Title: linkNode.Text(),
			ID:    findIntAndParse(link),
			Reply: findIntAndParse(node.Nodes[0].LastChild.Data),
		}
		threads = append(threads, thread)
	}
	return
}

// GetPosts returns posts of some thread at some page.
func (s *S1Client) GetPosts(thread Thread, page int) (posts []*Post, err error) {
	url := fmt.Sprintf(threadURLTemplate, thread.ID, page)
	doc, err := s.getAndParase(url)
	if err != nil {
		return
	}

	authorNodes := doc.Find(".author")
	timePattern := regexp.MustCompile("\\d+-\\d+-\\d+ \\d+:\\d+")
	authorPattern := regexp.MustCompile("<strong>(.*)</strong>")
	tzshanghai, _ := time.LoadLocation("Asia/Shanghai")
	for i := range authorNodes.Nodes {
		node := authorNodes.Eq(i)
		html, _ := node.Html()

		postTimeStr := timePattern.FindString(html)
		postTime, err := time.ParseInLocation("2006-01-02 15:04", postTimeStr, tzshanghai)
		if err != nil {
			return posts, err
		}
		author := authorPattern.FindStringSubmatch(html)[1]

		post := &Post{
			Author:   author,
			Content:  getPostContent(node),
			PostTime: postTime,
		}
		posts = append(posts, post)
	}
	return posts, nil
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

func getPostContent(author *goquery.Selection) (content string) {
	next := author.NextAllFiltered(".author,.page").Eq(0)
	buf := bytes.Buffer{}

	node := author.Nodes[0].NextSibling
	for ; node != nil && node != next.Nodes[0]; node = node.NextSibling {
		if node.Type == html.TextNode {
			buf.WriteString(node.Data)
		}
	}
	content = buf.String()

	// Remove \t s at begin/end of a line
	content = regexp.MustCompile("(?m:(^\\t+)|(\\t+$))").ReplaceAllString(content, "")
	// Remove new lines at begin/end of a post.
	content = regexp.MustCompile("(^[\\r\\n]+)|([\\r\\n]+)$").ReplaceAllString(content, "")
	return content
}
