package http

import (
	pb "charon/api"
	"charon/internal/model"
	"github.com/go-kratos/kratos/pkg/conf/paladin"
	"github.com/go-kratos/kratos/pkg/log"
	bm "github.com/go-kratos/kratos/pkg/net/http/blademaster"
	"net/http"
)

var svc pb.DemoServer

// New new a bm server.
func New(s pb.DemoServer) (engine *bm.Engine, err error) {
	var (
		cfg bm.ServerConfig
		ct  paladin.TOML
	)
	if err = paladin.Get("http.toml").Unmarshal(&ct); err != nil {
		return
	}
	if err = ct.Get("Server").UnmarshalTOML(&cfg); err != nil {
		return
	}
	svc = s
	//引擎
	engine = bm.DefaultServer(&cfg)
	//grpc-web服务
	pb.RegisterDemoBMServer(engine, s)
	//路由
	initRouter(engine)
	err = engine.Start()
	return
}

func initRouter(e *bm.Engine) {
	d := &Demo{}
	d2 := &Demo2{}
	cors := bm.CORS([]string{"github.com"})
	limiter := bm.NewRateLimiter(nil)
	e.Use(d, bm.Recovery(), bm.Trace(), bm.Logger(), cors, limiter.Limit())
	e.Ping(ping)
	g := e.Group("/charon")
	{
		g.GET("/start/:name", d2.ServeHTTP, howToStart)
	}
}

func ping(ctx *bm.Context) {
	if _, err := svc.Ping(ctx, nil); err != nil {
		log.Error("ping error(%v)", err)
		ctx.AbortWithStatus(http.StatusServiceUnavailable)
	}
	ctx.JSON(`pong`, nil)
}

// example for http request handler.
func howToStart(c *bm.Context) {
	var Msg string
	if name, ok := c.Params.Get(`name`); ok {
		Msg = `Hello ` + name + ` !`
	} else {
		Msg = "Golang 大法好 !!!"
	}
	k := &model.Kratos{
		Hello: Msg,
	}
	c.JSON(k, nil)
}

/*- 中间件1 -*/
type Demo struct {
}

func (d *Demo) ServeHTTP(ctx *bm.Context) {
	ctx.Set(`log`,9908877)
	log.Infoc(ctx,`test`)
}

/*- 中间件2 -*/
type Demo2 struct {
}

func (d *Demo2) ServeHTTP(ctx *bm.Context) {
	log.Infoc(ctx, `test-222`)
}
