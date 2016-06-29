package cache

import (
	"github.com/trihatmaja/goproxy"
	"gopkg.in/redis.v3"
	"log"
	"net/http"
	"net/http/httputil"
	"time"
)

func NewRedisCache(addr string, db int) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr: addr,
		DB:   int64(db),
	})
}

func RedisCacheRequestHandler(client *redis.Client) goproxy.ReqHandler {
	return goproxy.FuncReqHandler(func(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
		key := "url:" + req.RequestURI
		cmd := client.Get(key)
		v, _ := cmd.Result()

		if v == "" {
			return req, nil
		}

		return nil, cacheResponse(req, []byte(v))
	})
}

func RedisCacheResponseHandler(client *redis.Client) goproxy.RespHandler {
	return goproxy.FuncRespHandler(func(resp *http.Response, ctx *goproxy.ProxyCtx) *http.Response {
		key := "url:" + ctx.Req.RequestURI
		cmd := client.Get(key)
		v, _ := cmd.Result()
		if v == "" {
			respdump, err := httputil.DumpResponse(resp, true)
			if err != nil {
				errorResponse(ctx.Req)
			}
			cmd := client.Set(key, string(respdump), 300*time.Second)
			_, err = cmd.Result()
			if err != nil {
				errorResponse(ctx.Req)
			}
		}
		return resp
	})
}
