package goframe_prometheus

import (
	"net/http"

	"github.com/gogf/gf/v2/net/ghttp"
)

// PromAdapter wrappers the standard http.Handler to ghttp.HandlerFunc
func PromAdapter(handler http.Handler) ghttp.HandlerFunc {
	return ghttp.WrapH(handler)
}
