# goframe-prometheus
<p align="center">
    <em>Prometheus metrics exporter for goframe.</em>
</p>

该版本基于[ginprom](https://github.com/chenjiandongx/ginprom/#-ginprom)，适用于goframe框架

### 📝 Usage

```golang
package main

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	goframePrometheus "github.com/qiuyier/goframe-prometheus"
)

func main() {
	s := g.Server()

	// 注册metrics路由
	s.BindHandler("/metrics", goframePrometheus.PromAdapter(promhttp.Handler()))

	s.Group("/", func(group *ghttp.RouterGroup) {
		// 注册中间件，收集信息
		group.Middleware(goframePrometheus.PromMiddleWare)

		group.ALL("/hello", func(r *ghttp.Request) {
			r.Response.Write("hello, world!")
		})
	})

	s.SetPort(8000)
	s.Run()

}

```