package cache

import (
	"bytes"
	"io/ioutil"
	"net/http"
)

var errorMsg = []byte("500 Internal Proxy Error")

func errorResponse(req *http.Request) *http.Response {
	return &http.Response{
		StatusCode:    500,
		ProtoMajor:    1,
		ProtoMinor:    1,
		Request:       req,
		Header:        http.Header{"X-Proxy-Error": []string{"Internal Proxy Error"}},
		Body:          ioutil.NopCloser(bytes.NewBuffer(errorMsg)),
		ContentLength: int64(len(errorMsg)),
	}
}

func cacheResponse(req *http.Request, body []byte) *http.Response {
	return &http.Response{
		StatusCode:    200,
		ProtoMajor:    1,
		ProtoMinor:    1,
		Request:       req,
		Body:          ioutil.NopCloser(bytes.NewBuffer(body)),
		ContentLength: int64(len(body)),
	}
}
