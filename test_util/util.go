package test_util

import (
	"github.com/smy20011/s1go/test_util/data"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
)

const SinglePostThread = 111111

type MockS1Website struct {
}

func (m *MockS1Website) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	resp = &http.Response{
		Request: req,
	}

	url := req.URL.Path
	if strings.Contains(url, "111111") {
		resp.Body = createResponseBody("data/single.html")
	} else if strings.Contains(url, "fid") {
		resp.Body = createResponseBody("data/forum.html")
	} else if strings.Contains(url, "tid") {
		resp.Body = createResponseBody("data/thread.html")
	} else {
		resp.Body = createResponseBody("data/index.html")
	}
	return
}

func createResponseBody(file string) io.ReadCloser {
	reader := strings.NewReader(string(data.MustAsset(file)))
	return ioutil.NopCloser(reader)
}
