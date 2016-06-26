package cache

import (
	"bytes"
	gomemcache "github.com/bradfitz/gomemcache/memcache"
	"github.com/trihatmaja/goproxy"
	"gopkg.in/redis.v3"
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

func NewMemCache(s string) *gomemcache.Client {
	return gomemcache.New(s)
}

func NewRedisCache(addr string, db int) *redis.Client {
	return redis.NewClient(redis.Options{
		Addr: addr,
		DB:   int64(db),
	})
}

func MemCacheRequestHandler(client *gomemcache.Client) goproxy.ReqHandler {
	return goproxy.FuncReqHandler(func(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
		key := "url:" + req.RequestURI
		v, err := client.Get(key)
		if err != nil {
			return nil, errorResponse(req)
		}

		if v.Value == []byte("") {
			return req, nil
		}
		return nil, cacheResponse(req, v.Value)
	})
}

func MemCacheResponseHandler(client *gomemcache.Client) goproxy.RespHandler {
	return goproxy.FuncRespHandler(func(resp *http.Response, ctx *goproxy.ProxyCtx) *http.Response {
		key := "url:" + ctx.Req.RequestURI
		v, err := client.Get(key)
		if err != nil {
			return errorResponse(req)
		}

		if v.Value == []byte("") {
			myitem := &gomemcache.Item{
				Key:        key,
				Value:      resp.Body,
				Expiration: 300, // cache only for 5 minutes
			}
			err = client.Set(myitem)
			if err != nil {
				return errorResponse(req)
			}
		}
		return resp
	})
}

func RedisCacheRequestHandler(client *redis.Client) goproxy.ReqHandler {
	return goproxy.FuncReqHandler(func(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
		key := "url:" + req.RequestURI
		cmd := client.Get(key)
		v, err := cmd.Result()
		if err != nil {
			return nil, errorResponse(req)
		}
		if v == "" {
			return req, nil
		}
		return nil, cacheResponse(req, v)
	})
}

func RedisCacheResponseHandler(client *redis.Client) goproxy.RespHandler {
	return goproxy.FuncRespHandler(func(resp *http.Response, ctx *goproxy.ProxyCtx) *http.Response {
		key := "url:" + ctx.Req.RequestURI
		cmd := client.Get(key)
		v, err := cmd.Result()
		if err != nil {
			return errorResponse(req)
		}
		if v == "" {
			cmd := client.Set(key, resp.Body, 300)
			_, err = cmd.Result()
			if err != nil {
				errorResponse(req)
			}
		}
		return resp
	})
}

func LimitHttps(ratelimit *RateLimiter) goproxy.RespHandler {
	return goproxy.FuncHttpsHandler(func(host string, ctx *goproxy.ProxyCtx) (*goproxy.ConnectAction, string) {
		if ratelimit.Limit() {
			ctx.Resp = limiterResponse(ctx.Req)
			return goproxy.RejectConnect, host
		}
		return goproxy.OkConnect, host
	})
}
