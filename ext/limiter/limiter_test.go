package limiter_test

import (
	"crypto/tls"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/trihatmaja/goproxy"
	"github.com/trihatmaja/goproxy/ext/limiter"
)

var srv = httptest.NewServer(nil)
var https = httptest.NewTLSServer(nil)

var acceptAllCerts = &tls.Config{InsecureSkipVerify: true}

func oneShotProxy(proxy *goproxy.ProxyHttpServer, t *testing.T) (client *http.Client, s *httptest.Server) {
	s = httptest.NewServer(proxy)

	proxyUrl, _ := url.Parse(s.URL)
	tr := &http.Transport{TLSClientConfig: acceptAllCerts, Proxy: http.ProxyURL(proxyUrl)}
	client = &http.Client{Transport: tr}
	return
}

func getStatus(url string, client *http.Client) int {
	resp, _ := client.Get(url)
	return resp.StatusCode
}

func TestLimiterHttp(t *testing.T) {

	v := limiter.NewRateLimiter(1, time.Second)

	proxy := goproxy.NewProxyHttpServer()
	proxy.OnRequest().Do(limiter.LimitHttp(v))

	client, s := oneShotProxy(proxy, t)
	defer s.Close()

	r1 := getStatus(srv.URL, client)
	assert.Equal(t, http.StatusNotFound, r1, "they must be equal")

	r2 := getStatus(srv.URL, client)
	assert.Equal(t, http.StatusTooManyRequests, r2, "they must be equal")

	time.Sleep(1 * time.Second)

	r3 := getStatus(srv.URL, client)
	assert.Equal(t, http.StatusNotFound, r3, "they must be equal")

}
