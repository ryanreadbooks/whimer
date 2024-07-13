package xhttp

import (
	stdhttp "net/http"

	"github.com/zeromicro/go-zero/rest"
)

func Get(path string, handler stdhttp.HandlerFunc) rest.Route {
	return rest.Route{
		Method:  stdhttp.MethodGet,
		Path:    path,
		Handler: handler,
	}
}

func Post(path string, handler stdhttp.HandlerFunc) rest.Route {
	return rest.Route{
		Method:  stdhttp.MethodPost,
		Path:    path,
		Handler: handler,
	}
}

func Put(path string, handler stdhttp.HandlerFunc) rest.Route {
	return rest.Route{
		Method:  stdhttp.MethodPut,
		Path:    path,
		Handler: handler,
	}
}

func Delete(path string, handler stdhttp.HandlerFunc) rest.Route {
	return rest.Route{
		Method:  stdhttp.MethodDelete,
		Path:    path,
		Handler: handler,
	}
}

func Head(path string, handler stdhttp.HandlerFunc) rest.Route {
	return rest.Route{
		Method:  stdhttp.MethodHead,
		Path:    path,
		Handler: handler,
	}
}
