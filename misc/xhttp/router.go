package xhttp

import (
	"net/http"
	"path"

	"github.com/zeromicro/go-zero/rest"
)

// 对go-zero rest.Server的简单封装
// 分组写法和使用中间件更加方便一点
type RouterGroup struct {
	basePath    string
	core        *rest.Server
	middlewares []rest.Middleware
}

func NewRouterGroup(core *rest.Server) *RouterGroup {
	return &RouterGroup{
		basePath: "/",
		core:     core,
	}
}

func lastChar(str string) uint8 {
	if str == "" {
		panic("str is empty for lastChar")
	}
	return str[len(str)-1]
}

func joinPaths(absolutePath, relativePath string) string {
	if relativePath == "" {
		return absolutePath
	}

	finalPath := path.Join(absolutePath, relativePath)
	if lastChar(relativePath) == '/' && lastChar(finalPath) != '/' {
		return finalPath + "/"
	}
	return finalPath
}

func (g *RouterGroup) calculateAbsolutePath(relativePath string) string {
	return joinPaths(g.basePath, relativePath)
}

// 创建一个路由分组
func (g *RouterGroup) Group(path string, middlewares ...rest.Middleware) *RouterGroup {
	// 继承原先的中间件
	newMiddlewares := append(g.middlewares, middlewares...)
	ng := &RouterGroup{
		basePath:    g.calculateAbsolutePath(path),
		core:        g.core,
		middlewares: newMiddlewares,
	}

	return ng
}

func (g *RouterGroup) Use(middlewares ...rest.Middleware) {
	g.middlewares = append(g.middlewares, middlewares...)
}

func (g *RouterGroup) concatMiddlewares(middlewares ...rest.Middleware) []rest.Middleware {
	resMiddlewares := make([]rest.Middleware, 0, len(g.middlewares)+len(middlewares))
	resMiddlewares = append(resMiddlewares, g.middlewares...)
	resMiddlewares = append(resMiddlewares, middlewares...)

	return resMiddlewares
}

func (g *RouterGroup) Get(path string, handler http.HandlerFunc, middlewares ...rest.Middleware) {
	g.core.AddRoutes(
		rest.WithMiddlewares(g.concatMiddlewares(middlewares...), Get(path, handler)),
		rest.WithPrefix(g.basePath),
	)
}

func (g *RouterGroup) Post(path string, handler http.HandlerFunc, middlewares ...rest.Middleware) {
	g.core.AddRoutes(
		rest.WithMiddlewares(g.concatMiddlewares(middlewares...), Post(path, handler)),
		rest.WithPrefix(g.basePath),
	)
}

func (g *RouterGroup) Put(path string, handler http.HandlerFunc, middlewares ...rest.Middleware) {
	g.core.AddRoutes(
		rest.WithMiddlewares(g.concatMiddlewares(middlewares...), Put(path, handler)),
		rest.WithPrefix(g.basePath),
	)
}

func (g *RouterGroup) Delete(path string, handler http.HandlerFunc, middlewares ...rest.Middleware) {
	g.core.AddRoutes(
		rest.WithMiddlewares(g.concatMiddlewares(middlewares...), Delete(path, handler)),
		rest.WithPrefix(g.basePath),
	)
}

func (g *RouterGroup) Head(path string, handler http.HandlerFunc, middlewares ...rest.Middleware) {
	g.core.AddRoutes(
		rest.WithMiddlewares(g.concatMiddlewares(middlewares...), Head(path, handler)),
		rest.WithPrefix(g.basePath),
	)
}
