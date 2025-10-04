package remote

import (
	"net"
	"net/http"

	"github.com/ryanreadbooks/whimer/misc/metadata"
)

func ClientAddr(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rx := metadata.WithClientAddr(r.Context(), r.RemoteAddr)
		var clientIp string
		clientIp, _, _ = net.SplitHostPort(r.RemoteAddr) // ipv6-compatible
		rx = metadata.WithClientIp(rx, clientIp)
		next(w, r.WithContext(rx))
	}
}
