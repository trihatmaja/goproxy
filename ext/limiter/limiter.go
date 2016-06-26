package limiter

import (
	"bytes"
	"github.com/trihatmaja/goproxy"
	"io/ioutil"
	"net/http"
)

var limitedMsg = []byte("429 Too Many Request")

func limiterResponse(req *http.Request) *http.Response {
	return &http.Response{
		StatusCode:    429,
		ProtoMajor:    1,
		ProtoMinor:    1,
		Request:       req,
		Header:        http.Header{"X-Request-Limited": []string{"Too Many Request"}},
		Body:          ioutil.NopCloser(bytes.NewBuffer(limitedMsg)),
		ContentLength: int64(len(limitedMsg)),
	}
}

func LimitHttp(ratelimit *RateLimiter) goproxy.ReqHandler {
	return goproxy.FuncReqHandler(func(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
		if ratelimit.Limit() {
			return nil, limiterResponse(req)
		}
		return req, nil
	})
}

func LimitHttps(ratelimit *RateLimiter) goproxy.HttpsHandler {
	return goproxy.FuncHttpsHandler(func(host string, ctx *goproxy.ProxyCtx) (*goproxy.ConnectAction, string) {
		if ratelimit.Limit() {
			ctx.Resp = limiterResponse(ctx.Req)
			return goproxy.RejectConnect, host
		}
		return goproxy.OkConnect, host
	})
}
