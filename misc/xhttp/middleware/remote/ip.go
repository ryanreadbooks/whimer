package remote

import (
	"net/http"
	"strings"

	"github.com/ryanreadbooks/whimer/misc/metadata"
)

func ClientAddr(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rx := metadata.WithClientAddr(r.Context(), r.RemoteAddr)
		// TODO not ipv6 compatible
		res := strings.Split(r.RemoteAddr, ":")
		rx = metadata.WithClientIp(rx, res[0])

		next(w, r.WithContext(rx))
	}
}
