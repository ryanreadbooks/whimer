package cors

import (
	"net/http"
	"strings"

	"github.com/zeromicro/go-zero/rest"
)

// 跨域处理中间件
func Cors(origins []string) rest.Middleware {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			w.Header().Add("Access-Control-Allow-Origin", strings.Join(origins, ","))
			next(w, r)
		}
	}
}
func CorsAll(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Access-Control-Allow-Origin", "*")
		w.Header().Add("Access-Control-Allow-Credentials", "true")
		w.Header().Add("Access-Control-Allow-Methods", "*")
		next(w, r)
	}
}
