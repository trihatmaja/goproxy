package cache_test

import (
	"crypto/tls"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/trihatmaja/goproxy"
	"github.com/trihatmaja/goproxy/ext/cache"
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

func TestMemcache(t *testing.T) {
	c := cache.NewMemCache("127.0.0.1:11211")
	proxy := goproxy.NewProxyHttpServer()
	proxy.OnRequest().Do(cache.MemCacheRequestHandler(c))
	proxy.OnResponse().Do(cache.MemCacheResponseHandler(c))

	client, s := oneShotProxy(proxy, t)
	defer s.Close()

	r := getStatus(srv.URL, client)
	assert.Equal(t, http.StatusNotFound, r, "they must be equal")
}

func TestRediscache(t *testing.T) {
	c := cache.NewRedisCache("127.0.0.1:6379", 0)
	proxy := goproxy.NewProxyHttpServer()
	proxy.OnRequest().Do(cache.RedisCacheRequestHandler(c))
	proxy.OnResponse().Do(cache.RedisCacheResponseHandler(c))

	client, s := oneShotProxy(proxy, t)
	defer s.Close()

	r := getStatus(srv.URL, client)
	assert.Equal(t, http.StatusNotFound, r, "they must be equal")
}
