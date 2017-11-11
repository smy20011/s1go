package test_util

import (
	"github.com/smy20011/s1go/test_util/data"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
)

var contentMap = map[string]string{
	"fid": "data/forum.html",
	"tid": "data/thread.html",
}

type MockS1Website struct {
}

func (m *MockS1Website) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	resp = &http.Response{
		Request: req,
	}
	for key, file := range contentMap {
		if strings.Contains(req.URL.Path, key) {
			resp.Body = createResponseBody(file)
			return
		}
	}
	resp.Body = createResponseBody("data/index.html")
	return
}

func createResponseBody(file string) io.ReadCloser {
	reader := strings.NewReader(string(data.MustAsset(file)))
	return ioutil.NopCloser(reader)
}
