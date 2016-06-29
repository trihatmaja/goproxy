package cache

import (
	gomemcache "github.com/bradfitz/gomemcache/memcache"
	"github.com/trihatmaja/goproxy"
	// "log"
	// "fmt"
	"net/http"
	"net/http/httputil"
)

func NewMemCache(s string) *gomemcache.Client {
	return gomemcache.New(s)
}

func MemCacheRequestHandler(client *gomemcache.Client) goproxy.ReqHandler {
	return goproxy.FuncReqHandler(func(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
		key := "url:" + req.RequestURI
		v, err := client.Get(key)
		if err.Error() == "memcache: cache miss" {
			return req, nil
		}

		return nil, cacheResponse(req, v.Value)
	})
}

func MemCacheResponseHandler(client *gomemcache.Client) goproxy.RespHandler {
	return goproxy.FuncRespHandler(func(resp *http.Response, ctx *goproxy.ProxyCtx) *http.Response {
		key := "url:" + ctx.Req.RequestURI
		_, err := client.Get(key)
		if err != nil {
			respdump, err := httputil.DumpResponse(resp, true)
			if err != nil {
				return errorResponse(ctx.Req)
			}

			myitem := &gomemcache.Item{
				Key:        key,
				Value:      respdump,
				Expiration: 300, // cache only for 5 minutes
			}
			err = client.Set(myitem)
			if err != nil {
				return errorResponse(ctx.Req)
			}
		}

		return resp
	})
}
