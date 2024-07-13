package xhttp

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func TestRouter(t *testing.T) {
	engine, err := rest.NewServer(rest.RestConf{})
	if err != nil {
		t.Fatal(err)
	}
	router := RouterGroup{
		basePath: "/",
		core:     engine,
	}

	g1 := router.Group("/passport")
	g1.Get("/v1/me", nil)
	g1.Post("/v1/update", nil)

	router.Get("/v1/hello", nil)

	router.core.PrintRoutes()
}

func TestMiddleware(t *testing.T) {
	engine, err := rest.NewServer(rest.RestConf{})
	if err != nil {
		t.Fatal(err)
	}

	logMiddleware := func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			t.Log("before logging")
			next(w, r)
			t.Log("after logging")
			t.Log("--------------------")
		}
	}

	// 设置一些中间件
	router := RouterGroup{core: engine, basePath: "/"}
	router.Use(logMiddleware)

	{
		passportMiddleware := func(next http.HandlerFunc) http.HandlerFunc {
			return func(w http.ResponseWriter, r *http.Request) {
				t.Log("before g1")
				next(w, r)
				t.Log("after g1")
			}
		}
		g1 := router.Group("/passport", passportMiddleware)

		v3Middleware := func(next http.HandlerFunc) http.HandlerFunc {
			return func(w http.ResponseWriter, r *http.Request) {
				t.Log("v3 log before")
				next(w, r)
				t.Log("v3 log after")
			}
		}
		g1.Get("/v3/indie", func(w http.ResponseWriter, r *http.Request) {
			t.Log("v3 indie")
			httpx.Ok(w)
		}, v3Middleware)
		g1.Get("/v1/me", func(w http.ResponseWriter, r *http.Request) {
			t.Log("me")
			httpx.Ok(w)
		})
		g1.Post("/v2/update", func(w http.ResponseWriter, r *http.Request) {
			t.Log("update")
			httpx.Ok(w)
		})

	}

	router.core.PrintRoutes()

	ts := httptest.NewServer(engine)
	defer ts.Close()

	// 发起请求
	req1, _ := http.NewRequest(http.MethodGet, ts.URL+"/passport/v1/me", nil)
	http.DefaultClient.Do(req1)

	req2, _ := http.NewRequest(http.MethodPost, ts.URL+"/passport/v2/update", nil)
	http.DefaultClient.Do(req2)

	req3, _ := http.NewRequest(http.MethodGet, ts.URL+"/passport/v3/indie", nil)
	http.DefaultClient.Do(req3)
}
